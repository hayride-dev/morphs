#[allow(warnings)]
mod bindings;

use bindings::exports::hayride::datetime::datetime::Guest;

struct DateTime;

impl Guest for DateTime {
    fn date() -> String {
        let now = chrono::Utc::now();

        return now.format("%Y-%m-%d").to_string();
    }
}

bindings::export!(DateTime with_types_in bindings);