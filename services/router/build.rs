// services/router/build.rs
// fn main() -> Result<(), Box<dyn std::error::Error>> {
//     // Tell Cargo to rebuild if the proto file changes
//     println!("cargo:rerun-if-changed=/proto/router.proto");

//     tonic_build::configure()
//         .build_server(true)
//         .compile(
//             &["/proto/router.proto"], // Path inside the container
//             &["/proto"],               // Include path
//         )?;
//     Ok(())
// }

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. Better to use relative paths from the crate root
    // Let's assume your structure is: 
    // ├── proto/router.proto
    // └── services/router/build.rs
    let proto_file = "../../proto/router.proto";
    let proto_dir = "../../proto";

    // Tell Cargo to rebuild if the proto file changes
    println!("cargo:rerun-if-changed={}", proto_file);

    tonic_build::configure()
        .build_server(true)
        .compile(
            &[proto_file], 
            &[proto_dir], 
        )?;
    Ok(())
}