for i in {1..5} 
do
    ssh n$i killall -9 -w chaos-node
    ssh n$i rm -f /root/node.pid
done