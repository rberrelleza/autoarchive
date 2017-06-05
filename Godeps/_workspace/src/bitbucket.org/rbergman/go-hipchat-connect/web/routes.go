package web

// MountCommon mounts the following common routes:
// * GET    /descriptor            -> s.HandleDescriptor
// * GET    /healthcheck           -> s.HandleHealthCheck
// * POST   /installable           -> s.HandleInstall
// * DELETE /installable/:tenantID -> s.HandleUninstall
func (s *Server) MountCommon() {
	s.MountDescriptor()
	s.MountHealthCheck()
	s.MountInstallable()
}
