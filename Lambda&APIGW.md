# Lambda Part

## Prerequisite
* API Gateway for each Lambda Function
    * Method: *POST*
    * Method Request -> Request Body: 
    ```json
    {
        "ue_ID": "$input.params('ue_ID')"
    }
    ```

### Registration
```python
import boto3, json, os, datetime as dt
from random import randrange
from boto3.dynamodb.conditions import Key, Attr

def lambda_handler(event, context):
    # TODO implement
    body = json.loads(event['body'])
    ueID = body['ue_ID']
    dynamodb_resource = boto3.resource('dynamodb')
    table = dynamodb_resource.Table('Subscribe_User')

    startSession = table.put_item(
        Item={
            'ue-ID': ueID,
            'GU': 50
        }
    )
    
    output = ueID
    
    response = {'statusCode': 200, 'body' : output}
    
    return response
```

### Create
```python
import boto3, json, os, datetime as dt
from boto3.dynamodb.conditions import Key, Attr

def lambda_handler(event, context):
    body = json.loads(event['body'])
    ueID = body['ue_ID']
    dynamodb = boto3.resource('dynamodb', region_name='us-east-1')
    table = dynamodb.Table('Subscribe_User')
    dynamodb_client = boto3.client('dynamodb')
    existing_tables = dynamodb_client.list_tables()['TableNames']
    if ueID not in existing_tables:
        table = dynamodb.create_table(
            TableName= ueID,
            KeySchema=[
                {
                    'AttributeName': 'Event',
                    'KeyType': 'HASH'  # Partition key
                },
                {
                    'AttributeName': 'time',
                    'KeyType': 'RANGE'  # Sort key
                }
            ],
            AttributeDefinitions=[
                {
                    'AttributeName': 'Event',
                    'AttributeType': 'S'
                },
                {
                    'AttributeName': 'time',
                    'AttributeType': 'S'
                },
    
            ],
            ProvisionedThroughput={
                'ReadCapacityUnits': 10,
                'WriteCapacityUnits': 10
            }
        )
        
    output = ueID
            
    response = {'statusCode': 200, 'body' : output}
    return response
```

### Write Session
```python
import boto3, json, uuid, csv
from random import randrange
from boto3.dynamodb.conditions import Key, Attr

def lambda_handler(event, context):
    # TODO implement
    body = json.loads(event['body'])
    ueID = body['ue_ID']
    dynamodb_resource = boto3.resource('dynamodb')
    table = dynamodb_resource.Table(ueID)
    originalTable = dynamodb_resource.Table('Subscribe_User')
    originalUser = originalTable.query(
        KeyConditionExpression=Key('ue-ID').eq(ueID)
        )

    file_prefix = randrange(7)+1
    file_postfix = randrange(5)+1
    key = "sms-call-internet-mi-2013-11-0"+file_prefix+"-"+file_postfix +".csv"
    bucket = 'milan-dataset-fju'
    s3 = boto3.client('s3')
    
    confile= s3.get_object(Bucket=bucket, Key=key)
    recList = confile['Body'].read().decode().split('\n')
    firstrecord=True
    csv_reader = csv.reader(recList, delimiter=',', quotechar='"')
    for row in csv_reader:
        if (firstrecord):
            firstrecord=False
            continue
        datetime = row[0]
        CellID = row[1].replace(',','').replace('$','') if row[1] else '-'
        countrycode = row[2].replace(',','').replace('$','') if row[2] else '-'
        cost = randrange(10)+1
        smsin = row[3].replace(',','').replace('$','') if row[3] else '-'
        smsout = row[4].replace(',','').replace('$','') if row[4] else '-'
        callin = row[5].replace(',','').replace('$','') if row[5] else '-'
        callout = row[6].replace(',','').replace('$','') if row[6] else '-'
        internet = row[7].replace(',','').replace('$','') if row[7] else '-'

        startSession = table.put_item(
                    Item={
                        'Event' : uuid.uuid4().hex,
                        'time' : str(datetime),
                        'CellID' : CellID,
                        'countryCode' : countrycode,
                        'GU_Cost': cost,
                        'smsIn' : smsin,
                        'smsOut' : smsout,
                        'callIn' : callin,
                        'callOut' : callout,
                        'Internet' : internet
                    }
                )
        
    output = "Session continuing..."
    
    response = {'statusCode': 200, 'body' : output}
    
    return response
```

### Release
```python
import csv, boto3, json, os, datetime as dt
from boto3.dynamodb.conditions import Key, Attr

def lambda_handler(event, context):
    file_time = dt.datetime.fromtimestamp(os.path.getmtime(__file__))
    t = file_time.strftime("%Y-%m-%d-%H:%M:%S")
    
    #API Gateway
    body = json.loads(event['body'])
    ueID = body['ue_ID']
    
    #S3
    bucketName = ueID
    s3 = boto3.client('s3')
    s3.create_bucket(Bucket=bucketName)
    OUTPUT_BUCKET = bucketName
    TEMP_FILENAME = '/tmp/' + bucketName + "-" + t +'.csv'
    OUTPUT_KEY = bucketName + "-" + t +'.csv'
    s3_resource = boto3.resource('s3')
    
    #DynamoDB
    dynamodb_resource = boto3.resource('dynamodb')
    table = dynamodb_resource.Table(ueID)
    count = 0
    with open(TEMP_FILENAME, 'w') as output_file:
        writer = csv.writer(output_file)
        header = True
        first_page = True

        # Paginate results
        while True:

            # Scan DynamoDB table
            if first_page:
                response = table.scan()
                first_page = False
            else:
                response = table.scan(ExclusiveStartKey = response['LastEvaluatedKey'])

            for item in response['Items']:
                count = count + item['GU_Cost']
                # Write header row?
                if header:
                    writer.writerow(item.keys())
                    header = False

                writer.writerow(item.values())
                table.delete_item(
                    Key={
                        'Event': item['Event'],
                        'time': item['time']
                    }
                )

            # Last page?
            if 'LastEvaluatedKey' not in response:
                break

    # Upload temp file to S3
    s3_resource.Bucket(OUTPUT_BUCKET).upload_file(TEMP_FILENAME, OUTPUT_KEY)
    
    subscribertable = dynamodb_resource.Table('Subscribe_User')
    
    subscriberGU = subscribertable.query(
        KeyConditionExpression=Key('ue-ID').eq(ueID)
    )

    for i in subscriberGU['Items']:
        consumeGU = i['GU'] - count
        subscribertable.update_item(
            Key={
                'ue-ID': ueID,
            },
            UpdateExpression="set GU=:g",
            ExpressionAttributeValues={
                ':g': consumeGU
            },
            ReturnValues="UPDATED_NEW"
        )

    
    
    output = "Cost: " + str(count)
    
    response = {'statusCode': 200, 'body' : output}
    
    return response
```

### Update
```python
import boto3, json
from boto3.dynamodb.conditions import Key, Attr

def lambda_handler(event, context):
    body = json.loads(event['body'])
    ueID = body['ue_ID']
    dynamodb_resource = boto3.resource('dynamodb')
    table = dynamodb_resource.Table('Subscribe_User')
    
    response = table.query(
        KeyConditionExpression=Key('ue-ID').eq(ueID)
    )

    
    for i in response['Items']:
        if i['GU'] <= 0:
            table.update_item(
            Key={
                'ue-ID': ueID,
            },
            UpdateExpression="set GU=:g",
            ExpressionAttributeValues={
                ':g': 50
            },
            ReturnValues="UPDATED_NEW"
            )
            output = "GU Updated."
        else:
            output = "GU enough, no need to update."
    
    

    response = {'statusCode': 200, 'body' : output}
    return response
```
