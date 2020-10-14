
##


# slwh1 slwh2 slwh3  slwh4
# 104_slwh_1  104_slwh_2 104_slwh_3  104_slwh_4

#PROJECT_PATH="/usr/local/server/game/slwh/slwh1/bin"
#SUPER_NAME="104_slwh_1"

PROJECT_NAME="qp_104_slwh"
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
./upload.expect ../../bin/${PROJECT_NAME}_new ${USER_NAME} ${host} ${PASSWORD} /usr/local/server/game/slwh/slwh$1/bin
if [[ "$?" != 0 ]]; then
   exit 2
fi
done

echo '------------------restart-------------------'
# shellcheck disable=SC2068
for host in ${HOSTS[@]}
do
echo ${host}
./restart.expect ${PROJECT_NAME} ${USER_NAME} ${host} ${PASSWORD} /usr/local/server/game/slwh/slwh$1/bin 104_slwh_$1
done

rm -rf ../../bin/${PROJECT_NAME}_new