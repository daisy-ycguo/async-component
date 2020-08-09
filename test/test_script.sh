thetime=$(date +%s000)
echo "SCRIPT STARTED AT $thetime"
count=0
while [ $count -lt 10 ]
do
  curl "http://write-to-db.default.52.116.240.92.xip.io?duration=10&reqNum=$count" -H "Prefer: respond-async"
  count=`expr $count + 1`
done