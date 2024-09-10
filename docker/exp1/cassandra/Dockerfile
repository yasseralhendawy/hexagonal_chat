FROM cassandra:4.1


# Download MCAC
RUN curl -L -o /opt/mcac.tar.gz https://github.com/datastax/metric-collector-for-apache-cassandra/releases/download/v0.3.5/datastax-mcac-agent-0.3.5-4.1-beta1.tar.gz

# Unpack MCAC
RUN tar -xzf /opt/mcac.tar.gz -C /opt/ && mv /opt/datastax-mcac-agent-0.3.5-4.1-beta1 /opt/mcac

#Config MCAC 
RUN echo "JVM_OPTS=\"\$JVM_OPTS -javaagent:/opt/mcac/lib/datastax-mcac-agent.jar -Dcassandra.ring_delay_ms=0\"" >> /etc/cassandra/cassandra-env.sh

EXPOSE 9103
