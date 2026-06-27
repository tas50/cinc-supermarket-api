// Package supermarket is a Go client for the Chef Supermarket API.
//
// Read endpoints (cookbooks, search, tools, users, /universe, health) are
// anonymous; build a Client with no credentials:
//
//	c, _ := supermarket.NewClient(supermarket.Config{})
//	cb, _, _ := c.Cookbooks.Get(context.Background(), "apache2")
//
// Write endpoints (share, delete) use the Chef mixlib-authentication
// signed-header protocol (version 1.1, SHA-1 — the version the public
// Supermarket accepts). The credentials are the Supermarket username and the RSA
// private key whose public half is registered on that user's
// Supermarket profile:
//
//	key, _ := supermarket.LoadKeyFile("/home/me/.chef/me.pem")
//	c, _ := supermarket.NewClient(supermarket.Config{
//	    Username: "me",
//	    Key:      key,
//	})
//	c.Cookbooks.Share(ctx, "apache2", "Web Servers", tarball)
//
// All list and search endpoints share a paginated envelope; the
// generic Page[T] type carries the items along with Start/Items/Total
// for cursoring.
package supermarket
