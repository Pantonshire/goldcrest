use crate::twitter1;

impl From<i64> for twitter1::OptInt64 {
    fn from(x: i64) -> Self {
        twitter1::OptInt64{val: x}
    }
}

impl From<twitter1::OptInt64> for i64 {
    fn from(x: twitter1::OptInt64) -> Self {
        x.val
    }
}

impl From<u64> for twitter1::OptFixed64 {
    fn from(x: u64) -> Self {
        twitter1::OptFixed64{val: x}
    }
}

impl From<twitter1::OptFixed64> for u64 {
    fn from(x: twitter1::OptFixed64) -> Self {
        x.val
    }
}

impl From<u64> for twitter1::OptUint64 {
    fn from(x: u64) -> Self {
        twitter1::OptUint64{val: x}
    }
}

impl From<twitter1::OptUint64> for u64 {
    fn from(x: twitter1::OptUint64) -> Self {
        x.val
    }
}

impl From<String> for twitter1::OptString {
    fn from(x: String) -> Self {
        twitter1::OptString{val: x}
    }
}

impl From<twitter1::OptString> for String {
    fn from(x: twitter1::OptString) -> Self {
        x.val
    }
}
