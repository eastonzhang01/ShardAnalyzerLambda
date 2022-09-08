# OpenSearch Shard Analyzer Lambda Function

This tool analyzes the output of `_cat/shards?v` and recommend sharding strategy for the cluster based on the target shard size. 
### Guidelines
* **For search workloads** - target shard size could be 10-30 GB
* **For Log analytics workloads** - target shard size could be 30-50 GB
* By specifying the number of availability zones, the tool suggests well spread shards

### Note
This program was NOT created by me. I have only changed the main file to make this code work as a Lambda function. 

Previously, the tool has two modes: 'analyze' and 'server'.
With analyze, you input information using different flags in the command line.
With server, there will be a port for running the ShardAnalyzer as a server. It comes with swagger ui(`http://localhost:3000/swagger/index.html`) for easy endpoint testing. A report will be generated as a pdf.

As a lambda function, this program will take a JSON from APIGW and send back a response with the recommended sharding strategy. 
