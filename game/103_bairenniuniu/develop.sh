
##


# brnn1 brnn2 brnn3  brnn4
# 103_brnn_1  103_brnn_2 103_brnn_3  103_brnn_4

#PROJECT_PATH="/usr/local/server/game/brnn/brnn1/bin"
#SUPER_NAME="103_brnn_1"

PROJECT_NAME="qp_103_brnn"
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
./upload.expect ../../bin/${PROJECT_NAME}_new ${USER_NAME} ${host} ${PORT} ${PASSWORD} /usr/local/server/game/brnn/brnn$1/bin
if [[ "$?" != 0 ]]; then
   exit 2
fi
done

echo '------------------restart-------------------'
# shellcheck disable=SC2068
for host in ${HOSTS[@]}
do
echo ${host}
./restart.expect ${PROJECT_NAME} ${USER_NAME} ${host} ${PORT} ${PASSWORD} /usr/local/server/game/brnn/brnn$1/bin 103_brnn_$1
done

rm -rf ../../bin/${PROJECT_NAME}_new