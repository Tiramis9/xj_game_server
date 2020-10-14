
##

#scp -P 22022  -r bin keys config  root@47.56.172.167:/usr/local/server/api

# 101_0_lhd 101_1_lhd 101_2_lhd  101_3_lhd

PROJECT_PATH="/usr/local/server/hall_server/bin"
PROJECT_NAME="hall_server"
SUPER_NAME="hall"
USER_NAME="root"
HOSTS=("47.113.94.16")
PASSWORD="YC2JeVyZXWeXu3sT"
PORT="22"
#HOSTS=("216.118.243.18")
#PASSWORD="nFLLXiEopvpY"
#PORT="33888"


echo "Please Input the server password: "
#read -s PASSWORD

echo '------------------build------------------'
make linux
cp ../../bin/${PROJECT_NAME} ../../bin/${PROJECT_NAME}_new

echo '-----------------upload-----------------'
# shellcheck disable=SC2068
for host in ${HOSTS[@]}
do
echo ${host}
./upload.expect ../../bin/${PROJECT_NAME}_new ${USER_NAME} ${host} ${PORT} ${PASSWORD} ${PROJECT_PATH}
if [[ "$?" != 0 ]]; then
   exit 2
fi
done

echo '------------------restart-------------------'
# shellcheck disable=SC2068
for host in ${HOSTS[@]}
do
echo ${host}
./restart.expect ${PROJECT_NAME} ${USER_NAME} ${host} ${PORT} ${PASSWORD} ${PROJECT_PATH} ${SUPER_NAME}
done

rm -rf ../../bin/${PROJECT_NAME}_new