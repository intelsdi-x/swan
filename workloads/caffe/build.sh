set -e
mkdir -p output
echo "Build image ..."
docker build -t caffe .

echo "Check ..."
docker run --rm caffe /opt/swan/bin/caffe.sh --version

echo "Copy to output..."
docker run --rm -v $PWD/output:/output caffe cp -r /opt/swan/. /output
sudo chown -R $USER:$USER output

echo "Done."

echo "To install:" 
echo "sudo mkdir -p /opt/swan"
echo 'sudo chown -R $USER:$USER /opt/swan'
echo "cp -r output/* /opt/swan/"
echo "/opt/swan/bin/caffe.sh --version"
echo "/opt/swan/bin/caffe-test.sh"
