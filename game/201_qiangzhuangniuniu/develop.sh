
##


# qznn1 qznn2 qznn3 qznn4 qznn5
# 201_qznn_1  201_qznn_2 201_qznn_3  201_qznn_4 201_qznn_5

#PROJECT_PATH="/usr/local/server/game/qznn/qznn1/bin"
#SUPER_NAME="201_qznn_1"

PROJECT_NAME="qp_201_qznn"
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
./upload.expect ../../bin/${PROJECT_NAME}_new ${USER_NAME} ${host} ${PORT} ${PASSWORD} /usr/local/server/game/qznn/qznn$1/bin
if [[ "$?" != 0 ]]; then
   exit 2
fi
done

echo '------------------restart-------------------'
# shellcheck disable=SC2068
for host in ${HOSTS[@]}
do
echo ${host}
./restart.expect ${PROJECT_NAME} ${USER_NAME} ${host} ${PORT} ${PASSWORD} /usr/local/server/game/qznn/qznn$1/bin 201_qznn_$1
done

rm -rf ../../bin/${PROJECT_NAME}_new