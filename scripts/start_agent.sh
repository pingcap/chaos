for i in {1..5} 
do
    ssh n$i start-stop-daemon --start --background --make-pidfile --pidfile /root/agent.pid \
    --chdir /root --oknodo --startas /root/chaos-agent --name chaos-agent \
    -- --addr :8080 --log-file /root/agent.log
done