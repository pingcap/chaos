for i in {1..5} 
do
    ssh n$i start-stop-daemon --start --background --exec /root/chaos-node \
    --make-pidfile --pidfile /root/node.pid --chdir /root --oknodo --startas /root/chaos-node \
    -- --addr :8080
done