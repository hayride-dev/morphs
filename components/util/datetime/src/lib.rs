#[allow(warnings)]
mod bindings;

use bindings::exports::hayride::datetime::datetime::Guest;

struct DateTime;

impl Guest for DateTime {
    fn date() -> String {
        // Get the current date and time using local timezone
        let now = chrono::Local::now();
        let day_of_week = now.format("%A").to_string();
        let date = now.format("%B %d, %Y").to_string();
        let time = now.format("%I:%M %p").to_string();
        let timezone = now.format("%Z").to_string();

        return format!("Today's date is {}, {} and it is {} {}", day_of_week, date, time, timezone);
    }
}

bindings::export!(DateTime with_types_in bindings);