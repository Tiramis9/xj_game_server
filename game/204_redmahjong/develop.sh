
##


# zjh1 zjh2 zjh3  zjh4
# 202_zjh_1  202_zjh_2 202_zjh_3  202_zjh_4

#PROJECT_PATH="/usr/local/server/game/zjh/zjh1/bin"
#SUPER_NAME="202_zjh_1"

PROJECT_NAME="qp_204_hzmj"
USER_NAME="root"
HOSTS=("47.113.94.16")
PASSWORD="YC2JeVyZXWeXu3sT"

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
./upload.expect ../../bin/${PROJECT_NAME}_new ${USER_NAME} ${host} ${PASSWORD} /usr/local/server/game/hzmj/hzmj$1/bin
if [[ "$?" != 0 ]]; then
   exit 2
fi
done

echo '------------------restart-------------------'
# shellcheck disable=SC2068
for host in ${HOSTS[@]}
do
echo ${host}
./restart.expect ${PROJECT_NAME} ${USER_NAME} ${host} ${PASSWORD} /usr/local/server/game/hzmj/hzmj$1/bin 204_hzmj_$1
done

rm -rf ../../bin/${PROJECT_NAME}_new