#[allow(warnings)]
mod bindings;

pub mod datetime {
    pub use super::bindings::hayride::datetime;
}
