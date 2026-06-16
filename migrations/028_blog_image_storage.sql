-- Migration 028: Blog image storage bucket
-- Creates a Supabase Storage bucket for blog images with public read access

-- Create the storage bucket (run via Supabase Dashboard > Storage or SQL Editor)
INSERT INTO storage.buckets (id, name, public, file_size_limit, allowed_mime_types)
VALUES (
  'blog-images',
  'blog-images',
  true,
  5242880, -- 5MB limit
  ARRAY['image/png', 'image/jpeg', 'image/gif', 'image/webp', 'image/svg+xml']
)
ON CONFLICT (id) DO NOTHING;

-- Allow public read access to all blog images
CREATE POLICY "Public read access for blog images"
ON storage.objects FOR SELECT
USING (bucket_id = 'blog-images');

-- Allow authenticated users to upload blog images
CREATE POLICY "Authenticated users can upload blog images"
ON storage.objects FOR INSERT
WITH CHECK (
  bucket_id = 'blog-images'
  AND auth.role() = 'authenticated'
);

-- Allow authenticated users to delete their uploads
CREATE POLICY "Authenticated users can delete blog images"
ON storage.objects FOR DELETE
USING (
  bucket_id = 'blog-images'
  AND auth.role() = 'authenticated'
);
