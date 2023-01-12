use lockbook_core::{CancelSubscriptionError, PaymentMethod, PaymentPlatform, StripeAccountTier};

use crate::*;

#[repr(C)]
pub struct LbSubInfoResult {
    stripe_last4: *mut c_char,
    google_play_state: u8,
    app_store_state: u8,
    period_end: u64,
    err: LbError,
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_sub_info_result_free(r: LbSubInfoResult) {
    if !r.stripe_last4.is_null() {
        let _ = CString::from_raw(r.stripe_last4);
    }
    lb_error_free(r.err);
}

/// # Safety
///
/// The returned value must be passed to `lb_sub_info_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_subscription_info(core: *mut c_void) -> LbSubInfoResult {
    let mut r = LbSubInfoResult {
        stripe_last4: null_mut(),
        google_play_state: 0,
        app_store_state: 0,
        period_end: 0,
        err: lb_error_none(),
    };
    match core!(core).get_subscription_info() {
        Ok(None) => {} // Leave zero values for no subscription info.
        Ok(Some(info)) => {
            use PaymentPlatform::*;
            match info.payment_platform {
                Stripe { card_last_4_digits } => r.stripe_last4 = cstr(card_last_4_digits),
                // The integer representations of both the google play and app store account
                // state enums are bound together by a unit test in this crate.
                GooglePlay { account_state } => r.google_play_state = account_state as u8 + 1,
                AppStore { account_state } => r.app_store_state = account_state as u8 + 1,
            }
            r.period_end = info.period_end;
        }
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = LbErrorCode::Unexpected;
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_upgrade_account_stripe_old_card(core: *mut c_void) -> LbError {
    let mut e = lb_error_none();
    if let Err(err) =
        core!(core).upgrade_account_stripe(StripeAccountTier::Premium(PaymentMethod::OldCard))
    {
        e.msg = cstr(format!("{:?}", err));
        e.code = LbErrorCode::Unexpected;
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_upgrade_account_stripe_new_card(
    core: *mut c_void,
    num: *const c_char,
    exp_year: i32,
    exp_month: i32,
    cvc: *const c_char,
) -> LbError {
    let mut e = lb_error_none();
    let number = rstr(num).to_string();
    let cvc = rstr(cvc).to_string();

    if let Err(err) =
        core!(core).upgrade_account_stripe(StripeAccountTier::Premium(PaymentMethod::NewCard {
            number,
            exp_year,
            exp_month,
            cvc,
        }))
    {
        e.msg = cstr(format!("{:?}", err));
        e.code = LbErrorCode::Unexpected;
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_cancel_subscription(core: *mut c_void) -> LbError {
    let mut e = lb_error_none();
    if let Err(err) = core!(core).cancel_subscription() {
        use CancelSubscriptionError::*;
        e.msg = cstr(format!("{:?}", err));
        e.code = match err {
            Error::UiError(err) => match err {
                NotPremium => LbErrorCode::NotPremium,
                AlreadyCanceled => LbErrorCode::SubscriptionAlreadyCanceled,
                UsageIsOverFreeTierDataCap => LbErrorCode::UsageIsOverFreeTierDataCap,
                ExistingRequestPending => LbErrorCode::ExistingRequestPending,
                CouldNotReachServer => LbErrorCode::CouldNotReachServer,
                ClientUpdateRequired => LbErrorCode::ClientUpdateRequired,
                CannotCancelForAppStore => LbErrorCode::CannotCancelForAppStore,
            },
            Error::Unexpected(_) => LbErrorCode::Unexpected,
        };
    }
    e
}
