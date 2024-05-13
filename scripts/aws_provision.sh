#!/bin/bash

# AWS_REGION="us-east-2"
AWS_REGION="ap-southeast-1"
PLACEMENT_GROUP_NAME="boki-experiments"
SECURITY_GROUP_NAME="boki"
IAM_ROLE_NAME="boki-ae-experiments"

awscmd="aws --output text --region $AWS_REGION"

LOCAL_IP=$(ip addr | grep 'state UP' -A2 | tail -n1 | awk '{print $2}' | cut -f1  -d'/')
if [ -z "$LOCAL_IP" ]; then
    echo "Cannot determine local IP address"
    exit 1
fi

# Create placement group
if [[ $($awscmd ec2 describe-placement-groups --group-names $PLACEMENT_GROUP_NAME 2>/dev/null) ]]; then
    echo "Reusing existing placement group $PLACEMENT_GROUP_NAME"
else
    echo "Cannot access placement group $PLACEMENT_GROUP_NAME, trying to create a new one"
    $awscmd ec2 create-placement-group --group-name $PLACEMENT_GROUP_NAME --strategy cluster
fi

# Remove existing security group
if [[ $($awscmd ec2 describe-security-groups --group-names $SECURITY_GROUP_NAME 2>/dev/null) ]]; then
    echo "Removing existing security group $SECURITY_GROUP_NAME"
    $awscmd ec2 delete-security-group --group-name $SECURITY_GROUP_NAME
fi

# Create security group
echo "Trying to create new security group $SECURITY_GROUP_NAME"
SECURITY_GROUP_ID=$($awscmd ec2 create-security-group \
    --group-name $SECURITY_GROUP_NAME --description "Boki experiments")

if [ -z "$SECURITY_GROUP_ID" ]; then
    echo "Failed to create security group $SECURITY_GROUP_NAME"
    exit 1
fi

# Allow all internal traffic within the security group
$awscmd ec2 authorize-security-group-ingress \
    --group-id $SECURITY_GROUP_ID \
    --ip-permissions "IpProtocol=-1,FromPort=-1,ToPort=-1,UserIdGroupPairs=[{GroupId=$SECURITY_GROUP_ID}]"

# Allow SSH traffic from current machine to the security group
$awscmd ec2 authorize-security-group-ingress \
    --group-id $SECURITY_GROUP_ID \
    --ip-permissions "IpProtocol=tcp,FromPort=22,ToPort=22,IpRanges=[{CidrIp=$LOCAL_IP/32}]"

# Set up IAM role and attach policy for DynamoDB access
if [[ $($awscmd iam get-instance-profile --instance-profile-name $IAM_ROLE_NAME 2>/dev/null) ]]; then
    echo "Reusing existing IAM role $IAM_ROLE_NAME"
else
    echo "Cannot access IAM role $IAM_ROLE_NAME, trying to create a new one"
    $awscmd iam create-role \
        --role-name $IAM_ROLE_NAME \
        --assume-role-policy-document file://trust_relationships.json \
        --description "Grant DynamoDB full access to EC2 service" \
        --max-session-duration 43200

    $awscmd iam attach-role-policy \
        --role-name $IAM_ROLE_NAME \
        --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess

    # Associate an instace profile with it
    $awscmd iam create-instance-profile \
        --instance-profile-name $IAM_ROLE_NAME

    $awscmd iam add-role-to-instance-profile \
        --instance-profile-name $IAM_ROLE_NAME \
        --role-name $IAM_ROLE_NAME
fi

