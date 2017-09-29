for i in {1..5} 
do
    ssh n$i start-stop-daemon --stop --name chaos-agent --pidfile /root/agent.pid --oknodo --remove-pidfile
done