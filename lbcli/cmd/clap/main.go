package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/structtag"
	"golang.org/x/tools/go/packages"
)

// todo:
// - ruleset: disallow bool args

var helpOption = optInfo{
	short: "h",
	long:  "help",
	desc:  "show this help message",
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	rootCmdName := flag.String("name", "", "root command name (default's to given type name)")
	rootCmdTypeName := flag.String("type", "", "root command struct type (required)")
	flag.Parse()

	fset := token.NewFileSet() // positions are relative to fset
	parsedDir, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgLoadCfg := packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(&pkgLoadCfg, ".")
	if err != nil {
		return fmt.Errorf("loading package: %w", err)
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("no packages loaded")
	}

	g := generator{
		pkg: parsedDir["main"],
		out: &bytes.Buffer{},
	}

	fmt.Fprintf(g.out, outFileHeader, "%s", "%s")

	var obj types.Object
	pkgDefs := pkgs[0].TypesInfo.Defs
	for i := range pkgDefs {
		if _, ok := pkgDefs[i].(*types.TypeName); ok {
			if *rootCmdTypeName == pkgDefs[i].Name() {
				obj = pkgDefs[i]
				break
			}
		}
	}
	if obj == nil {
		return fmt.Errorf("did not find type '%s'", *rootCmdTypeName)
	}

	strct, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return fmt.Errorf("root command type must be struct, '%s' is a '%T'", *rootCmdTypeName, obj.Type().Underlying())
	}

	name := *rootCmdName
	if name == "" {
		name = strings.ToLower(obj.Name())
	}
	desc := g.lookupTypeComments(*rootCmdTypeName)
	if desc == nil {
		warn("no root command description provided\n")
	}
	if err := g.genCommandCode(cmdMetadata{
		parentNames:    []string{},
		strct:          strct,
		typeName:       *rootCmdTypeName,
		usageConstName: *rootCmdTypeName + "Usage",
		name:           name,
		desc:           desc,
	}); err != nil {
		return fmt.Errorf("generating root command code: %w", err)
	}

	f, err := os.OpenFile("./clap.go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("opening clap out file: %v", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, g.out); err != nil {
		return fmt.Errorf("copying to clap out file: %w", err)
	}
	return nil
}

type cmdMetadata struct {
	parentNames    []string
	strct          *types.Struct
	name           string
	desc           *ast.CommentGroup
	typeName       string
	usageConstName string
	usageArgs      string
	subcmds        []subcmdInfo
	opts           []optInfo
	args           []argInfo
}

type subcmdInfo struct {
	fieldName string
	typeName  string
	name      string
	desc      string
}

type optInfo struct {
	basicType *types.Basic
	fieldName string
	short     string
	long      string
	desc      string
}

type argInfo struct {
	basicType *types.Basic
	fieldName string
	name      string
	desc      string
	required  bool
}

type generator struct {
	pkg *ast.Package
	out *bytes.Buffer
}

func (g *generator) lookupTypeComments(typ string) *ast.CommentGroup {
	var commentGrp *ast.CommentGroup
	ast.Inspect(g.pkg, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.GenDecl:
			if n.Doc != nil {
				for i := range n.Specs {
					if s, ok := n.Specs[i].(*ast.TypeSpec); ok && s.Name.Name == typ {
						commentGrp = n.Doc
						return false
					}
				}
			}
		case *ast.TypeSpec:
			if n.Name.Name == typ && n.Doc != nil {
				commentGrp = n.Doc
				return false
			}
		}
		return true
	})
	return commentGrp
}

func (g *generator) genCommandCode(m cmdMetadata) error {
	if m.hasArgs() && m.hasSubcmds() {
		return fmt.Errorf("commands cannot contain both arguments and subcommands, '%s' has both", m.typeName)
	}
	// Read each struct field to determine if it's an option, argument, or subcommand.
	for i := 0; i < m.strct.NumFields(); i++ {
		field := m.strct.Field(i)
		tags, err := structtag.Parse(m.strct.Tag(i))
		if err != nil {
			return fmt.Errorf("unable to parse struct tags: %v", err)
		}
		// Basic types are either options (if they contain the 'opt' tag key) or arguments.
		if _, ok := field.Type().(*types.Basic); ok {
			otag, err := tags.Get("opt")
			if err != nil || otag == nil {
				m.buildArg(field, tags)
				continue
			}
			if otag.Name == "" {
				return fmt.Errorf("%s.%s tag value for 'opt' cannot be empty", m.typeName, field.Name())
			}
			m.buildOption(field, tags, otag)
			continue
		}
		// Subcommands must be named struct pointers.
		if ptr, ok := field.Type().(*types.Pointer); ok {
			typeName := typeNameBase(ptr.Elem().(*types.Named).String())
			subCmdMeta := cmdMetadata{
				parentNames:    append(m.parentNames, m.name),
				strct:          ptr.Elem().Underlying().(*types.Struct),
				typeName:       typeName,
				usageConstName: typeName + "Usage",
				usageArgs:      tagsGetValue(tags, "uargs"),
				name:           field.Name(),
				desc:           g.lookupTypeComments(typeName),
			}
			m.buildSubCmd(&subCmdMeta)
			if err := g.genCommandCode(subCmdMeta); err != nil {
				return err
			}
			continue
		}
		// Print out a warning if this field is being skipped, plus a helpful extra one if
		// it's because there was a plain struct and not a pointer.
		warn("skipping %s.%s (not an option, argument, or subcommand)\n", m.typeName, field.Name())
		if _, ok := field.Type().(*types.Struct); ok {
			fmt.Printf("%s.%s is a plain struct, must be a pointer\n", m.typeName, field.Name())
		}
		continue
	}
	m.opts = append(m.opts, helpOption)

	g.writeCmdUsageMsg(&m)
	g.writeCmdParseFunc(&m)
	return nil
}

func (m *cmdMetadata) buildOption(field types.Object, tags *structtag.Tags, otag *structtag.Tag) {
	long := ""
	short := ""
	switch len(otag.Options) {
	case 0:
		if len(otag.Name) == 1 {
			short = otag.Name
		} else {
			long = otag.Name
		}
	case 1:
		if len(otag.Name) == 1 {
			long, short = otag.Options[0], otag.Name
		} else if len(otag.Options[0]) == 1 {
			long, short = otag.Name, otag.Options[0]
		} else {
			log.Fatalf("two opt names found ('%s', '%s'), one must be the short version (only one character)", otag.Name, otag.Options[0])
		}
	default:
		log.Fatalf("illegal `opt` tag value '%s': too many comma separated values", otag)
	}
	if long == "help" || short == "h" {
		log.Fatal("'help' and 'h' are reserved option names")
	}
	oi := optInfo{
		basicType: field.Type().(*types.Basic),
		fieldName: field.Name(),
		short:     short,
		long:      long,
		desc:      tagsGetValue(tags, "desc"),
	}
	m.opts = append(m.opts, oi)
}

func (m *cmdMetadata) buildArg(field types.Object, tags *structtag.Tags) {
	ai := argInfo{
		basicType: field.Type().(*types.Basic),
		fieldName: field.Name(),
		name:      strings.ToLower(field.Name()),
	}
	if atag, err := tags.Get("arg"); err == nil || atag != nil {
		ai.desc = atag.Name
		ai.required = atag.HasOption("required")
	}
	m.args = append(m.args, ai)
}

func (m *cmdMetadata) buildSubCmd(subCmd *cmdMetadata) {
	ci := subcmdInfo{
		fieldName: subCmd.name,
		typeName:  subCmd.typeName,
		name:      subCmd.name,
		desc:      subCmd.shortDesc(),
	}
	if ci.name == "" {
		ci.name = strings.ToLower(ci.fieldName)
	}
	m.subcmds = append(m.subcmds, ci)
}

func (m *cmdMetadata) shortDesc() string {
	if m.desc == nil {
		return ""
	}
	c := m.desc.Text()
	if c[len(c)-1] == '\n' {
		c = c[:len(c)-1]
	}
	return c
}

func (g *generator) writeCmdUsageMsg(m *cmdMetadata) {
	fmt.Fprintf(g.out, "\nconst %s = `", m.usageConstName)

	parents := strings.Join(m.parentNames, " ")
	if parents != "" {
		parents += " "
	}
	fmt.Fprintf(g.out, "name:\n   %s%s - %s\n\n", parents, m.name, m.shortDesc())

	if m.usageArgs != "" {
		fmt.Fprintf(g.out, "usage:\n   %s %s\n", m.name, m.usageArgs)
	} else {
		optionsSlot := " [options]" // Every command has at least the help options for now.
		commandSlot := ""
		if m.hasSubcmds() {
			commandSlot = " <command>"
		}
		argsSlot := ""
		if m.hasArgs() {
			for _, ai := range m.args {
				argsSlot += " " + ai.docString()
			}
		}
		fmt.Fprintf(g.out, "usage:\n   %s%s%s%s\n", m.name, optionsSlot, commandSlot, argsSlot)
	}

	if m.hasSubcmds() {
		cmdNameWidth := 0
		for _, ci := range m.subcmds {
			if len(ci.name) > cmdNameWidth {
				cmdNameWidth = len(ci.name)
			}
		}
		cmdNameWidth++

		fmt.Fprintf(g.out, "\ncommands:\n")
		for _, ci := range m.subcmds {
			fmt.Fprintf(g.out, "   %-*s %s\n", cmdNameWidth, ci.name, ci.desc)
		}
	}

	if m.hasArgs() {
		argNameWidth := 0
		for _, ai := range m.args {
			if len(ai.name) > argNameWidth {
				argNameWidth = len(ai.docString())
			}
		}
		argNameWidth += 2

		fmt.Fprintf(g.out, "\narguments:\n")
		for _, ai := range m.args {
			fmt.Fprintf(g.out, "   %-*s %s\n", argNameWidth, ai.docString(), ai.desc)
		}
	}

	if m.hasOptions() {
		fmt.Fprintf(g.out, "\noptions:\n")
		namesWidth := m.optNamesColWidth()
		for _, oi := range m.opts {
			fmt.Fprintf(g.out, "   %-*s  %s\n", namesWidth, oi.docNames(), oi.desc)
		}
	}
}

func (m *cmdMetadata) optNamesColWidth() int {
	w := 0
	for _, oi := range m.opts {
		if n := len(oi.docNames()); n > w {
			w = n
		}
	}
	return w + 1
}

func (g *generator) writeCmdParseFunc(m *cmdMetadata) {
	fmt.Fprintf(g.out, "`\n\nfunc (c *%s) parse(args []string) {\n", m.typeName)

	// Parse options.
	fmt.Fprintf(g.out, `	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
`)
	if m.hasOptField() {
		fmt.Fprintln(g.out, "\t\tname, eqv := optParts(args[i][1:])")
	} else {
		fmt.Fprintln(g.out, "\t\tname, _ := optParts(args[i][1:])")
	}
	fmt.Fprintln(g.out, "\t\tswitch name {")
	for _, opt := range m.opts {
		fmt.Fprintf(g.out, "\t\tcase %s:\n", opt.quotedPlainNames())
		// Hard code the 'help' case.
		if opt.long == "help" {
			fmt.Fprintf(g.out, "\t\t\texitUsgGood(%s)\n", m.usageConstName)
			continue
		}
		switch opt.basicType.Kind() {
		case types.Bool:
			fmt.Fprintf(g.out, "\t\t\tc.%s = (eqv == \"\" || eqv == \"true\")\n", opt.fieldName)
		}
	}
	fmt.Fprint(g.out, "\t\t}\n") // end switch
	fmt.Fprint(g.out, "\t}\n")   // end loop

	// Arguments.
	if m.hasArgs() {
		fmt.Fprint(g.out, "\targs = args[i:]\n")
		// Add error handling for missing arguments that are required.
		reqArgs := m.requiredArgs()
		for i := range reqArgs {
			fmt.Fprintf(g.out, "\tif len(args) < %d {\n", i+1)
			fmt.Fprintf(g.out, "\t\texitMissingArg(%q, %s)\n", reqArgs[i].docString(), m.usageConstName)
			fmt.Fprint(g.out, "\t}\n")
		}
		for i, arg := range m.args {
			if !arg.required {
				fmt.Fprintf(g.out, "\tif len(args) < %d {\n", i+1)
				fmt.Fprint(g.out, "\t\treturn\n")
				fmt.Fprint(g.out, "\t}\n")
			}
			// Parse positional args based on their type.
			switch arg.basicType.Kind() {
			case types.String:
				fmt.Fprintf(g.out, "\tc.%s = args[%d]\n", arg.fieldName, i)
			default:
				log.Fatalf("%d arg types are not supported yet", arg.basicType.Kind())
			}
		}
	}

	// Subcommands.
	if m.hasSubcmds() {
		fmt.Fprint(g.out, "\tif i >= len(args) {\n")
		fmt.Fprintf(g.out, "\t\tfmt.Fprint(os.Stderr, %s)\n", m.usageConstName)
		fmt.Fprint(g.out, "\t\tos.Exit(1)\n")
		fmt.Fprint(g.out, "\t}\n")

		fmt.Fprint(g.out, "\tswitch args[i] {\n")

		for _, ci := range m.subcmds {
			fmt.Fprintf(g.out, "\tcase %q:\n", ci.name)
			fmt.Fprintf(g.out, "\t\tc.%s = new(%s)\n", ci.fieldName, ci.typeName)
			fmt.Fprintf(g.out, "\t\tc.%s.parse(args[i+1:])\n", ci.fieldName)
		}

		// Default care which means an unknown command.
		fmt.Fprint(g.out, "\tdefault:\n")
		fmt.Fprintf(g.out, "\t\texitUnknownCmd(args[i], %s)\n", m.usageConstName)
		fmt.Fprint(g.out, "\t}\n") // end switch
	}

	fmt.Fprintf(g.out, "}\n") // Closing curly for parse func.
}

func (m *cmdMetadata) requiredArgs() []argInfo {
	reqs := make([]argInfo, 0, len(m.args))
	for _, arg := range m.args {
		if arg.required {
			reqs = append(reqs, arg)
		}
	}
	return reqs
}

func (oi *optInfo) quotedPlainNames() string {
	long := oi.long
	if long != "" {
		long = "\"" + long + "\""
	}
	short := oi.short
	if short != "" {
		short = "\"" + short + "\""
	}
	comma := ""
	if oi.long != "" && oi.short != "" {
		comma = ", "
	}
	return fmt.Sprintf("%s%s%s", long, comma, short)
}

func (oi *optInfo) docNames() string {
	long := oi.long
	if long != "" {
		long = "--" + long
	}
	short := oi.short
	if short != "" {
		short = "-" + short
	}
	comma := ""
	if oi.long != "" && oi.short != "" {
		comma = ", "
	}
	return fmt.Sprintf("%s%s%s", long, comma, short)
}

func (ai *argInfo) docString() string {
	delims := []string{"[", "]"}
	if ai.required {
		delims = []string{"<", ">"}
	}
	return delims[0] + ai.name + delims[1]
}

func typeNameFromObj(obj types.Object) string {
	name := ""
	switch t := obj.Type().Underlying().(type) {
	case *types.Struct:
		name = obj.Type().(*types.Named).String()
	case *types.Pointer:
		name = t.Elem().(*types.Named).String()
	default:
		panic(fmt.Sprintf("%s is a %T, must be struct or pointer", obj, t))
	}
	return name[strings.LastIndexByte(name, '.')+1:]
}

func typeNameBase(name string) string {
	return name[strings.LastIndexByte(name, '.')+1:]
}

func tagsGetValue(tags *structtag.Tags, k string) string {
	t, err := tags.Get(k)
	if err != nil {
		return ""
	}
	return t.Name
}

func (m *cmdMetadata) hasOptField() bool {
	for _, o := range m.opts {
		if o.long != "help" {
			return true
		}
	}
	return false
}

func (m *cmdMetadata) hasSubcmds() bool { return len(m.subcmds) > 0 }
func (m *cmdMetadata) hasOptions() bool { return len(m.opts) > 0 }
func (m *cmdMetadata) hasArgs() bool    { return len(m.args) > 0 }

func warn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "\033[1;33mwarning:\033[0m "+format, a...)
}

const outFileHeader = `// This file is generated via 'go generate'; DO NOT EDIT
package main

import (
	"fmt"
	"os"
	"strings"
)

func claperr(format string, a ...any) {
	format = "\033[1;31merror:\033[0m " + format
	fmt.Fprintf(os.Stderr, format, a...)
}

func exitEmptyOpt() {
	claperr("emtpy option ('-') found\n")
	os.Exit(1)
}

func exitMissingArg(name, u string) {
	claperr("not enough args: no \033[1;33m%s\033[0m provided\n", name)
	fmt.Fprint(os.Stderr, u)
	os.Exit(1)
}

func exitUnknownCmd(name, u string) {
	claperr("unknown command '%s'\n", name)
	fmt.Fprint(os.Stderr, u)
	os.Exit(1)
}

func exitUsgGood(u string) {
	fmt.Fprint(os.Stdout, u)
	os.Exit(0)
}

func optParts(arg string) (string, string) {
	if arg == "-" {
		exitEmptyOpt()
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	name := arg
	eqVal := ""
	if eqIdx := strings.IndexByte(name, '='); eqIdx != -1 {
		name = arg[:eqIdx]
		eqVal = arg[eqIdx+1:]
	}
	return name, eqVal
}
`
