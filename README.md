# igotyou

**a Geo-index based social network**

*Users can send, search posts based on location*

- Adopted OAuth (JSON Web Token) for user registration and verification to reduce database query
- Built a web service with Go to handle posts and deployed it to Google Cloud (GAE flex) for load balancing
- Deployed Elasticsearch (GCE) to provide efficient location-based search such that users can search nearby posts 
- Stored large-scale (PB) structured data for cloud computing in Bigtable
- Implemented a daily dump of posts via Google Dataflow to BigQuery for offline analysis
- Aggregated the data at the post level and user level to improve the keyword-based spam detection (BigQuery)
- Trained the model by Cloud Machine Learning API for better prediction such as Face Detection (TensorFlow) 
- Reduced the response time by Redis Cloud as cache