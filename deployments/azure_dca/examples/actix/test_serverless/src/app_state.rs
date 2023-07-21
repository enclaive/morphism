use std::env;

pub(crate) struct AppState {
    pub(crate) message:String,
}

impl AppState {
    pub(crate) async fn generate() -> Self {
        let args: Vec<String> = env::args().collect();
        //let message = fs::read_to_string("/tmp/msg")
        //.expect("Should have been able to read the file");
        let mut message = "not enough arguments supplied".to_string();
        if args.len()>0{
            message = args[args.len()-1].clone()
        }
        AppState {
            message
        }
    }
}
