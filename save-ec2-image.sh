#!/bin/bash
#
# Creates an EC2 image.

# Variables that you need to configure...
AWS_ACCOUNT_ID=076857755310
AWS_ACCESS_KEY=TODO
AWS_SECRET=TODO
ZONE=ap
REGION=ap-northeast-1
AMI_BUCKET_NAME=dcooneyamis
VOLUME=vol-1b232d71
export EC2_KEYPAIR=WebKitTokyo
export EC2_PRIVATE_KEY=/mnt/tmp/WebKitTokyo.pem
export EC2_CERT=/mnt/tmp/cert-K756465P4PQ6RA4F2NZ2VIO6UN7S3TAM.pem

# ...variables that you don't need to configure.
DISTRIBUTION=$(lsb_release -a | awk '/Codename/ {print $2}')
export EC2_URL=https://ec2.$REGION.amazonaws.com
export JAVA_HOME=/usr/lib/jvm/java-6-openjdk/

sudo cat >> /etc/apt/sources.list <<EOF
deb http://$ZONE.archive.ubuntu.com/ubuntu/ $DISTRIBUTION universe
deb-src http://$ZONE.archive.ubuntu.com/ubuntu/ $DISTRIBUTION universe
deb http://$ZONE.archive.ubuntu.com/ubuntu/ $DISTRIBUTION-updates universe
deb-src http://$ZONE.archive.ubuntu.com/ubuntu/ $DISTRIBUTION-updates universe
deb http://$ZONE.archive.ubuntu.com/ubuntu/ $DISTRIBUTION multiverse
deb-src http://$ZONE.archive.ubuntu.com/ubuntu/ $DISTRIBUTION multiverse
deb http://$ZONE.archive.ubuntu.com/ubuntu/ $DISTRIBUTION-updates multiverse
deb-src http://$ZONE.archive.ubuntu.com/ubuntu/ $DISTRIBUTION-updates multiverse
EOF

sudo apt-get update
sudo apt-get install ec2-api-tools

# See if it works...
ec2-describe-images -o self -o amazon

sudo ec2-bundle-vol --fstab /etc/fstab -e /$VOLUME -d /mnt -k $EC2_PRIVATE_KEY -c $EC2_CERT -u $AWS_ACCOUNT_ID -r x86_64
ec2-upload-bundle -b $AMI_BUCKET_NAME --location $REGION -m /mnt/image.manifest.xml -a $AWS_ACCESS_KEY -s $AWS_SECRET
ec2-register -a x86_64 -U http://$REGION.ec2.amazonaws.com $AMI_BUCKET_NAME/image.manifest.xml
