package main

func sanitizeRoot(root string) string {
	if root == "" {
		return "/"
	}
	if root[0] != '/' {
		return "/" + root
	}
	if root[len(root)-1] == '/' {
		return root[:len(root)-1]
	}
	return root
}

func sanitizeUri(uri string) string {
	if uri == "" {
		return "/"
	}
	if uri[0] != '/' {
		return "/" + uri
	}
	if uri[len(uri)-1] == '/' {
		return uri[:len(uri)-1]
	}
	return uri
}
