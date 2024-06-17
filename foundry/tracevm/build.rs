
fn main() {
    let lib_path = ::std::env::var("DEP_PATH").expect("Please provide the `DEP_PATH` env var");

    println!("cargo::rustc-link-search={}", lib_path);
    println!("cargo::rustc-link-lib=dep");
}
