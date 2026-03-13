{
  cpu: 1,
  memory: 1,
  image: {
    path: "ghcr.io/fujiwara/printenv",
    tag: "v0.2.5",
  },
  exposed_ports: [
    {
      target_port: 8080,
      load_balancer_port: 80,
    },
  ],
}
