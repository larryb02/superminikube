/*
API between superminikube and container runtime (docker)
 */
package superminikube.containers;

import java.lang.ProcessBuilder;

class Container {

    private String id;
    private String imageName;
    private boolean status; // TODO: Create Enum

    public Container() {
        
    }
    public void Start() {
    }

    public void Stop() {

    }
}

public class ContainerRuntime {
    private static final String DOCKER_CMD = "docker";
    public static void Run(){
        ProcessBuilder pb = new ProcessBuilder(DOCKER_CMD);
        try {
            Process p = pb.start();
        } catch (Exception e) {
            // TODO: handle exception
            System.out.println(e);
        }
        
    }
    public void Stop(){}
    public void Pull(){}

    public static void main(String[] args) {
        System.out.println("It's starting...");
        ContainerRuntime.Run();
        System.out.println("It successfully ended.");
    }
}
