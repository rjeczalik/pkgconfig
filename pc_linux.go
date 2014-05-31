package pkgconfig

func init() {
	defaultPaths = append(defaultPaths,
		"/usr/lib/pkgconfig",
		"/usr/share/pkgconfig",
		"/usr/local/lib/pkgconfig",
		"/usr/local/share/pkgconfig",
	)
}
