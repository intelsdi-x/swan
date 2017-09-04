if [ "$EUID" -ne 0 ]
	then echo "Please run this setup as root"
	return
fi

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]
	then echo "Please run this setup with 'source' command"
	exit
fi


### Install required packages
echo "Installing required packages..."

# git
apt install git -y

# golang
bash local_scripts/goinstall.sh --64
source ~/.bashrc

# caffe


### Install and configure caffe


### Get latest Swan repository
git clone https://github.com/intelsdi-x/swan.git

### Build Swan
mkdir -p $GOPATH/src/github.com/intelsdi-x/swan
mv swan $GOPATH/src/github.com/intelsdi-x/
cd $GOPATH/src/github.com/intelsdi-x/swan/
make build_and_test_unit
cd -
