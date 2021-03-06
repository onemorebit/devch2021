package main

const freebsdCform = `---
AWSTemplateFormatVersion: '2010-09-09'
Description: 'FreeBSD Stack'

Resources:
  ResVPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16

  ResIGW:
    Type: AWS::EC2::InternetGateway

  AttachmentVPCandIGW:
    Type: AWS::EC2::VPCGatewayAttachment
    Properties:
      InternetGatewayId:
        Ref: ResIGW
      VpcId:
        Ref: ResVPC

  ResRTIGW:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId:
        Ref: ResVPC

  AttachmentRoutingToIGW:
    Type: AWS::EC2::Route
    DependsOn:
      - ResIGW
      - ResRTIGW
    Properties:
      RouteTableId:
        Ref: ResRTIGW
      DestinationCidrBlock: 0.0.0.0/0
      GatewayId:
        Ref: ResIGW


  ResSUBNET:
    Type: AWS::EC2::Subnet
    DependsOn: AttachmentRoutingToIGW
    Properties:
      VpcId:
        Ref: ResVPC
      CidrBlock:
        Fn::GetAtt:
          - ResVPC
          - CidrBlock

  AssociationSNandRT:
    Type: AWS::EC2::SubnetRouteTableAssociation
    DependsOn: ResSUBNET
    Properties:
      SubnetId:
        Ref: ResSUBNET
      RouteTableId:
        Ref: ResRTIGW
  ResEC2SG:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Enable public SSH access
      VpcId:
        Ref: ResVPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: '22'
          ToPort: '22'
          CidrIp: '0.0.0.0/0'

  ResIAMRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: ec2.amazonaws.com
            Action: sts:AssumeRole
      Path: "/"
      ManagedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AmazonEC2RoleforSSM

  ResEC2Profile:
    Type: AWS::IAM::InstanceProfile
    Properties:
      Path: "/"
      Roles:
        - Ref: ResIAMRole

  ResEC2:
    Type: AWS::EC2::Instance
    Properties:
      InstanceType: "t2.small"
      KeyName: "alpha"
      ImageId: "ami-063de173ab8f5bfd9" # custom image based on ami-0762f1426163a4437 with the preinstalled packages x11/kde5 sysutils/amazon-ssm-agent
      IamInstanceProfile: !Ref ResEC2Profile
      BlockDeviceMappings:
        - DeviceName: "/dev/sda1"
          Ebs:
            VolumeSize: '16'
            VolumeType: gp2
      NetworkInterfaces:
        - SubnetId:
            Ref: ResSUBNET
          DeviceIndex: '0'
          AssociatePublicIpAddress: 'true'
          DeleteOnTermination: 'true'
          GroupSet:
            - Ref: ResEC2SG
# see ImageId: comments
#      UserData:
#        Fn::Base64: !Sub |
#          #!/bin/sh
#          su root -c "pkg install -y sysutils/amazon-ssm-agent"
#          su root -c "pkg install -y x11/kde5"
#          su root -c 'echo amazon_ssm_agent_enable=YES >> /etc/rc.conf'
#          su root -c 'service amazon-ssm-agent start'
#          su root -c 'rm -rf  /etc/ssh/ssh_host_*'

Outputs:
  InstanceId:
    Description: InstanceId
    Value:
      Ref: ResEC2

  PublicIP:
    Description: EC2 Public IP address
    Value:
      Fn::GetAtt:
        - ResEC2
        - PublicIp
`
