# Quick Deployment on AWS

* [free5gc EC2](#free5gc-ec2)
* [free5gc ECS](#free5gc-ecs)

## free5gc EC2

```yaml
AWSTemplateFormatVersion: 2010-09-09
Parameters:
  InstanceTypeParameter:
    Type: String
    Default: t2.xlarge
    Description: Enter instance size. Default is t2.xlarge.
    AllowedValues: [t2.medium, t2.xlarge, t2.2xlarge, t3.xlarge, t3.2xlarge]
  AMI:
    Type: String
    Default: ami-0b13a869090f3c0be
    Description: The free5gc environment pre-setup as Ubuntu 20.04.
  Key:
    Description: Name of an existing EC2 KeyPair to enable SSH access to the instance
    Type: AWS::EC2::KeyPair::KeyName
    ConstraintDescription: must be the name of an existing EC2 KeyPair.
Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      EnableDnsSupport: true
      EnableDnsHostnames: true
      InstanceTenancy: default
      Tags:
        - Key: Name
          Value: Free5gc VPC
  InternetGateway:
    Type: AWS::EC2::InternetGateway
  VPCGatewayAttachment:
    Type: AWS::EC2::VPCGatewayAttachment
    Properties:
      VpcId: !Ref VPC
      InternetGatewayId: !Ref InternetGateway
  PublicSubnet:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone: us-east-1a
      VpcId: !Ref VPC
      CidrBlock: 10.0.0.0/24
      MapPublicIpOnLaunch: true
  RouteTable:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
  InternetRoute:
    Type: AWS::EC2::Route
    DependsOn: InternetGateway
    Properties:
      DestinationCidrBlock: 0.0.0.0/0
      GatewayId: !Ref InternetGateway
      RouteTableId: !Ref RouteTable
  PublicSubnetRouteTableAssociation:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref RouteTable
      SubnetId: !Ref PublicSubnet
  InstanceSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupName: "Internet Group"
      GroupDescription: "SSH traffic in, all traffic out."
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: "-1"
          CidrIp: 0.0.0.0/0
  ElasticIP:
    Type: AWS::EC2::EIP
    Properties:
      Domain: vpc
      InstanceId: !Ref free5gc
  free5gc:
    Type: 'AWS::EC2::Instance'
        
    Properties:
      SubnetId: !Ref PublicSubnet
      ImageId: !Ref AMI
      InstanceType:
        Ref: InstanceTypeParameter
      KeyName: !Ref Key
      SecurityGroupIds:
        - Ref: InstanceSecurityGroup
      BlockDeviceMappings:
        - DeviceName: /dev/sda1
          Ebs:
            SnapshotId: snap-0075598523b0e7d9d
            VolumeSize: 160
      Tags:
        -
          Key: Name
          Value: Free5gc Server
      UserData:
        Fn::Base64: |
          #!/bin/bash
          cd /home/ubuntu/free5gcWithOCF
          for i in {1..10}; do sudo -u ubuntu sudo ./test.sh TestRegistration; done
         #100 UEs
         #for i in {1..100}; do sudo -u ubuntu sudo ./test.sh TestRegistration; done
         #1000 UEs
         #for i in {1..1000}; do sudo -u ubuntu sudo ./test.sh TestRegistration; done
         #10000 UEs
         #for i in {1..10000}; do sudo -u ubuntu sudo ./test.sh TestRegistration; done
         #100000 UEs
         #for i in {1..100000}; do sudo -u ubuntu sudo ./test.sh TestRegistration; done
          
Outputs:
  InstanceId:
    Description: InstanceId of the newly created EC2 instance
    Value: !Ref 'free5gc'
  PublicIP:
    Description: Public IP address of the newly created EC2 instance
    Value: !GetAtt [free5gc, PublicIp]
```

## free5gc ECS

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: EC2 ECS cluster running containers in a public subnet. Only supports
             public facing load balancer, and public service discovery namespaces.
Parameters:
  EnvironmentName:
    Type: String
    Default: production
    Description: "A friendly environment name that will be used for namespacing all cluster resources. Example: staging, qa, or production"
  InstanceTypeParameter:
    Type: String
    Default: t2.xlarge
    Description: Enter instance size. Default is t2.xlarge.
    AllowedValues: [t2.medium, t2.xlarge, t2.2xlarge, t3.xlarge, t3.2xlarge]
  DesiredCapacity:
    Type: Number
    Default: '1'
    Description: Number of EC2 instances to launch in your ECS cluster.
  MaxSize:
    Type: Number
    Default: '3'
    Description: Maximum number of EC2 instances that can be launched in your ECS cluster.
  ECSAMI:
    Type: String
    Default: ami-01fb9b7c7c4320711
    Description: The free5gc environment pre-setup as Ubuntu 20.04.

Mappings:
  SubnetConfig:
    VPC:
      CIDR: '10.0.0.0/16'
    PublicOne:
      CIDR: '10.0.0.0/24'
    PublicTwo:
      CIDR: '10.0.1.0/24'
Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      EnableDnsSupport: true
      EnableDnsHostnames: true
      CidrBlock: !FindInMap ['SubnetConfig', 'VPC', 'CIDR']
  PublicSubnetOne:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone:
         Fn::Select:
         - 0
         - Fn::GetAZs: {Ref: 'AWS::Region'}
      VpcId: !Ref 'VPC'
      CidrBlock: !FindInMap ['SubnetConfig', 'PublicOne', 'CIDR']
      MapPublicIpOnLaunch: true
  PublicSubnetTwo:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone:
         Fn::Select:
         - 1
         - Fn::GetAZs: {Ref: 'AWS::Region'}
      VpcId: !Ref 'VPC'
      CidrBlock: !FindInMap ['SubnetConfig', 'PublicTwo', 'CIDR']
      MapPublicIpOnLaunch: true
  InternetGateway:
    Type: AWS::EC2::InternetGateway
  GatewayAttachement:
    Type: AWS::EC2::VPCGatewayAttachment
    Properties:
      VpcId: !Ref 'VPC'
      InternetGatewayId: !Ref 'InternetGateway'
  PublicRouteTable:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref 'VPC'
  PublicRoute:
    Type: AWS::EC2::Route
    DependsOn: GatewayAttachement
    Properties:
      RouteTableId: !Ref 'PublicRouteTable'
      DestinationCidrBlock: '0.0.0.0/0'
      GatewayId: !Ref 'InternetGateway'
  PublicSubnetOneRouteTableAssociation:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      SubnetId: !Ref PublicSubnetOne
      RouteTableId: !Ref PublicRouteTable
  PublicSubnetTwoRouteTableAssociation:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      SubnetId: !Ref PublicSubnetTwo
      RouteTableId: !Ref PublicRouteTable
  ECSCluster:
    Type: AWS::ECS::Cluster
  ContainerSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Access to the ECS hosts that run containers
      VpcId: !Ref 'VPC'
  ECSAutoScalingGroup:
    Type: AWS::AutoScaling::AutoScalingGroup
    Properties:
      VPCZoneIdentifier:
        - !Ref PublicSubnetOne
        - !Ref PublicSubnetTwo
      LaunchConfigurationName: !Ref 'ContainerInstances'
      MinSize: '1'
      MaxSize: !Ref 'MaxSize'
      DesiredCapacity: !Ref 'DesiredCapacity'
    CreationPolicy:
      ResourceSignal:
        Timeout: PT15M
    UpdatePolicy:
      AutoScalingReplacingUpdate:
        WillReplace: 'true'
  ContainerInstances:
    Type: AWS::AutoScaling::LaunchConfiguration
    Properties:
      ImageId: !Ref 'ECSAMI'
      SecurityGroups: [!Ref 'ContainerSecurityGroup']
      InstanceType: !Ref 'InstanceType'
      IamInstanceProfile: !Ref 'EC2InstanceProfile'
      UserData:
        Fn::Base64: !Sub |
          #!/bin/bash -xe
          echo ECS_CLUSTER=${ECSCluster} >> /etc/ecs/ecs.config
          yum install -y aws-cfn-bootstrap
          /opt/aws/bin/cfn-signal -e $? --stack ${AWS::StackName} --resource ECSAutoScalingGroup --region ${AWS::Region}
          cd /home/ubuntu/free5gcWithOCF
          sudo -u ubuntu sudo ./test.sh TestRegistration
  EC2InstanceProfile:
    Type: AWS::IAM::InstanceProfile
    Properties:
      Path: /
      Roles: [!Ref 'EC2Role']
  AutoscalingRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Statement:
        - Effect: Allow
          Principal:
            Service: [application-autoscaling.amazonaws.com]
          Action: ['sts:AssumeRole']
      Path: /
      Policies:
      - PolicyName: service-autoscaling
        PolicyDocument:
          Statement:
          - Effect: Allow
            Action:
              - 'application-autoscaling:*'
              - 'cloudwatch:DescribeAlarms'
              - 'cloudwatch:PutMetricAlarm'
              - 'ecs:DescribeServices'
              - 'ecs:UpdateService'
            Resource: '*'
  EC2Role:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Statement:
        - Effect: Allow
          Principal:
            Service: [ec2.amazonaws.com]
          Action: ['sts:AssumeRole']
      Path: /
      Policies:
      - PolicyName: ecs-service
        PolicyDocument:
          Statement:
          - Effect: Allow
            Action:
              - 'ecs:CreateCluster'
              - 'ecs:DeregisterContainerInstance'
              - 'ecs:DiscoverPollEndpoint'
              - 'ecs:Poll'
              - 'ecs:RegisterContainerInstance'
              - 'ecs:StartTelemetrySession'
              - 'ecs:Submit*'
              - 'logs:CreateLogStream'
              - 'logs:PutLogEvents'
              - 'ecr:GetAuthorizationToken'
              - 'ecr:BatchGetImage'
              - 'ecr:GetDownloadUrlForLayer'
            Resource: '*'
  ECSRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Statement:
        - Effect: Allow
          Principal:
            Service: [ecs.amazonaws.com]
          Action: ['sts:AssumeRole']
      Path: /
      Policies:
      - PolicyName: ecs-service
        PolicyDocument:
          Statement:
          - Effect: Allow
            Action:
              - 'ec2:AttachNetworkInterface'
              - 'ec2:CreateNetworkInterface'
              - 'ec2:CreateNetworkInterfacePermission'
              - 'ec2:DeleteNetworkInterface'
              - 'ec2:DeleteNetworkInterfacePermission'
              - 'ec2:Describe*'
              - 'ec2:DetachNetworkInterface'
              - 'elasticloadbalancing:DeregisterInstancesFromLoadBalancer'
              - 'elasticloadbalancing:DeregisterTargets'
              - 'elasticloadbalancing:Describe*'
              - 'elasticloadbalancing:RegisterInstancesWithLoadBalancer'
              - 'elasticloadbalancing:RegisterTargets'
            Resource: '*'
Outputs:
  ClusterName:
    Description: The name of the ECS cluster
    Value: !Ref 'ECSCluster'
    Export:
      Name: !Sub ${EnvironmentName}:ClusterName
  AutoscalingRole:
    Description: The ARN of the role used for autoscaling
    Value: !GetAtt 'AutoscalingRole.Arn'
    Export:
      Name: !Sub ${EnvironmentName}:AutoscalingRole
  ECSRole:
    Description: The ARN of the ECS role
    Value: !GetAtt 'ECSRole.Arn'
    Export:
      Name: !Sub ${EnvironmentName}:ECSRole
  VpcId:
    Description: The ID of the VPC that this stack is deployed in
    Value: !Ref 'VPC'
    Export:
      Name: !Sub ${EnvironmentName}:VpcId
  PublicSubnetOne:
    Description: Public subnet one
    Value: !Ref 'PublicSubnetOne'
    Export:
      Name: !Sub ${EnvironmentName}:PublicSubnetOne
  PublicSubnetTwo:
    Description: Public subnet two
    Value: !Ref 'PublicSubnetTwo'
    Export:
      Name: !Sub ${EnvironmentName}:PublicSubnetTwo
  ContainerSecurityGroup:
    Description: A security group used to allow containers to receive traffic
    Value: !Ref 'ContainerSecurityGroup'
    Export:
      Name: !Sub ${EnvironmentName}:ContainerSecurityGroup

```