for i in {1..5} 
do
    ssh n$i start-stop-daemon --stop --name chaos-node --pidfile /root/node.pid --oknodo
    ssh n$i rm -f /root/node.pid
done