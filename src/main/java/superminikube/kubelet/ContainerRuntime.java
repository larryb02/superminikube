/*
API between superminikube and container runtime (docker)
 */
package superminikube.kubelet;

import java.time.Duration;

import com.github.dockerjava.api.DockerClient;
import com.github.dockerjava.core.DockerClientConfig;
import com.github.dockerjava.core.DockerClientImpl;
import com.github.dockerjava.httpclient5.ApacheDockerHttpClient;
import com.github.dockerjava.core.DefaultDockerClientConfig;
import com.github.dockerjava.transport.DockerHttpClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

class Container {

    private String id;
    // private String imageName; // think about this later -> would be useful for
    // some optimization techniques...
    // not needed at the moment and not easily accessible without a spec file
    private boolean status; // TODO: Create Enum
}

public class ContainerRuntime {
    final Logger logger = LoggerFactory.getLogger(ContainerRuntime.class);
    DockerClient dockerClient;

    public ContainerRuntime() {
        DockerClientConfig config = DefaultDockerClientConfig.createDefaultConfigBuilder()
                .withDockerHost("unix:///var/run/docker.sock")
                .build();
        DockerHttpClient httpClient = new ApacheDockerHttpClient.Builder().dockerHost(config.getDockerHost())
                .sslConfig(config.getSSLConfig())
                .maxConnections(100)
                .connectionTimeout(Duration.ofSeconds(30))
                .responseTimeout(Duration.ofSeconds(45))
                .build();
        this.dockerClient = DockerClientImpl.getInstance(config, httpClient);
    }

    public void Ping() {
        try {
            this.dockerClient.pingCmd().exec();
        } catch (Exception e) {
            // TODO: handle exception
            logger.error(e.toString());
        }
    }

    public void Create() {
        try {
            this.dockerClient.createContainerCmd(null).exec();
        } catch (Exception e) {
            // TODO: handle exception
        }
    }

    public void Stop() {
    }

    public void Pull() {
    }

    public static void main(String[] args) {
        ContainerRuntime cr = new ContainerRuntime();
        cr.Ping();
    }
}
