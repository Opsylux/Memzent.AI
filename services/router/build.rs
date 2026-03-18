// services/router/build.rs
fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Tell Cargo to rebuild if the proto file changes
    println!("cargo:rerun-if-changed=/proto/router.proto");

    tonic_build::configure()
        .build_server(true)
        .compile(
            &["/proto/router.proto"], // Path inside the container
            &["/proto"],               // Include path
        )?;
    Ok(())
}