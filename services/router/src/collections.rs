use qdrant_client::Qdrant;
use qdrant_client::qdrant::{
    CreateCollectionBuilder, Distance, ScalarQuantizationBuilder, VectorParamsBuilder,
    OptimizersConfigDiff, CreateFieldIndexCollectionBuilder, FieldType,
};

/// Initializes a Qdrant collection with SQ quantization, on-disk payloads,
/// memmap threshold, and payload field indexes for org_id and user_id.
pub async fn init_optimized_collection(
    q_client: &Qdrant,
    collection_name: &str,
    exists: bool,
) -> Result<(), Box<dyn std::error::Error>> {
    if !exists {
        println!("🚀 Creating optimized collection: {}", collection_name);

        q_client
            .create_collection(
                CreateCollectionBuilder::new(collection_name)
                    .vectors_config(VectorParamsBuilder::new(384, Distance::Cosine))
                    .quantization_config(ScalarQuantizationBuilder::default())
                    .on_disk_payload(true)
                    .optimizers_config(OptimizersConfigDiff {
                        memmap_threshold: Some(20000),
                        ..Default::default()
                    })
            )
            .await?;

        println!("🔑 Creating payload indexes for {}", collection_name);
        let _ = q_client.create_field_index(
            CreateFieldIndexCollectionBuilder::new(collection_name, "org_id", FieldType::Keyword)
        ).await;

        let _ = q_client.create_field_index(
            CreateFieldIndexCollectionBuilder::new(collection_name, "user_id", FieldType::Keyword)
        ).await;
    }
    Ok(())
}
