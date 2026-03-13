{
  cluster: "default",
  name: "printenv",
  cpu: 1,
  memory: 1,
  image: "ghcr.io/fujiwara/printenv:v0.2.5",
  exposedPorts: [
    {
      targetPort: 8080,
      loadBalancerPort: 80,
      host: ["app.example.com"],
    },
  ],
}
