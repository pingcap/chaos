for i in {1..5} 
do
    ssh n$i start-stop-daemon --start --background --make-pidfile --pidfile /root/node.pid \
    --chdir /root --oknodo --startas /root/chaos-node --name chaos-node \
    -- --addr :8080 --log-file /root/node.log
done