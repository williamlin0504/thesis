# UE Simulation

## Prerequisite
* Local Python Script
* Lambda Python Script
* API Gateway

## Local Python Script
* Add Credential to local ***(access key, secret access key)***

```python
import boto3

client = boto3.client('lambda')

#Set up i UE
# range(1,11) 10 
# range(1,101) 100 
# range(1,1001) 1000 
# range(1,10001) 10000
for i in range(1,10):
    createFunction = client.create_function(
        FunctionName='<Function-Name>',
        Runtime='python3.8',
        Role='<Role-ARN>',
        Handler='lambda_handler',
        Code={
        'S3Bucket':'<S3-Bucket-Name>', 
        'S3Key':'<File-Name>'
        },
        Description='<Description>',
        Layers=[
        '<Lambda-Layers-ARN>',
        ],
        Timeout=900,
        MemorySize=128
    )

    invokeFunction = client.invoke(
    FunctionName='<Function-Name>'
    )
```

## Lambda Python Script

* Upload to S3 Bucket

```python
import json, requests

def lambda_handler(event, context):
    # TODO implement
    response = requests.post(
        '<API URL>', 
        data = {'key':'value'}
    )
    return {
        'statusCode': 200,
        'body': json.dumps('Hello from Lambda!')
    }
```