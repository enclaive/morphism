use serde::{Serialize};

#[derive(Serialize)]
pub(crate) struct Response{
   pub(crate) message:String,
}
 
