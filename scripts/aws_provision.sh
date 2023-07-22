#!/bin/bash

# AWS_REGION="us-east-2"
AWS_REGION="ap-southeast-1"
PLACEMENT_GROUP_NAME="boki-experiments"
SECURITY_GROUP_NAME="boki"
IAM_ROLE_NAME="boki-ae-experiments"

# Create placement group
aws --output text --region $AWS_REGION ec2 create-placement-group \
    --group-name $PLACEMENT_GROUP_NAME --strategy cluster

# Create security group
SECURITY_GROUP_ID=$(\
    aws --output text --region $AWS_REGION ec2 create-security-group \
    --group-name $SECURITY_GROUP_NAME --description "Boki experiments")

# Allow all internal traffic within the newly create security group
aws --output text --region $AWS_REGION ec2 authorize-security-group-ingress \
    --group-id $SECURITY_GROUP_ID \
    --ip-permissions "IpProtocol=-1,FromPort=-1,ToPort=-1,UserIdGroupPairs=[{GroupId=$SECURITY_GROUP_ID}]"

LOCAL_IP=$(ip addr | grep 'state UP' -A2 | tail -n1 | awk '{print $2}' | cut -f1  -d'/')
# LOCAL_IP=$(ip addr | grep 'state UP' -A3 | tail -n1 | awk '{print $2}' | cut -f1  -d'/')
echo $LOCAL_IP

# Allow SSH traffic from current machine to the newly create security group
aws --output text --region $AWS_REGION ec2 authorize-security-group-ingress \
    --group-id $SECURITY_GROUP_ID \
    --ip-permissions "IpProtocol=tcp,FromPort=22,ToPort=22,IpRanges=[{CidrIp=$LOCAL_IP/32}]"

# Set up IAM role and attach policy for DynamoDB access
aws --output text --region $AWS_REGION iam create-role \
    --role-name $IAM_ROLE_NAME \
    --assume-role-policy-document file://trust_relationships.json \
    --description "Grant DynamoDB full access to EC2 service" \
    --max-session-duration 43200

aws --output text --region $AWS_REGION iam attach-role-policy \
    --role-name $IAM_ROLE_NAME \
    --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess

# Associate an instace profile with it
aws --output text --region $AWS_REGION iam create-instance-profile \
    --instance-profile-name $IAM_ROLE_NAME

aws --output text --region $AWS_REGION iam add-role-to-instance-profile \
    --instance-profile-name $IAM_ROLE_NAME \
    --role-name $IAM_ROLE_NAME

