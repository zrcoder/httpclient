package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

func (c *Client) TlsConfig(config *tls.Config) *Client {
	c.transport.TLSClientConfig = config
	return c
}

func (c *Client) InsecureSkipVerify(skip bool) *Client {
	c.judge2genTlsConfig()
	c.transport.TLSClientConfig.InsecureSkipVerify = skip
	return c
}

func (c *Client) AddCAFile(cafile string) *Client {
	content, err := ioutil.ReadFile(cafile)
	if err != nil {
		c.keepOrigionErr(err)
		return c
	}
	return c.AddCAContent(content)
}

func (c *Client) AddCAContent(cacontent []byte) *Client {
	c.judge2genPool()
	c.transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(cacontent)
	return c
}

func (c *Client) AddCACert(cert *x509.Certificate) *Client {
	c.judge2genPool()
	c.transport.TLSClientConfig.RootCAs.AddCert(cert)
	return c
}

func (c *Client) CertPool(pool *x509.CertPool) *Client {
	c.judge2genTlsConfig()
	c.transport.TLSClientConfig.RootCAs = pool
	return c
}

func (c *Client) AddCertFile(cert, key string) *Client {
	certCotent, err := ioutil.ReadFile(cert)
	if err != nil {
		c.keepOrigionErr(err)
		return c
	}
	keyContent, err := ioutil.ReadFile(key)
	if err != nil {
		c.keepOrigionErr(err)
		return c
	}
	return c.AddCertContent(certCotent, keyContent)
}

func (c *Client) AddCertContent(certContent, keyContent []byte) *Client {
	cert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		c.keepOrigionErr(err)
		return c
	}
	return c.AddCert(cert)
}

func (c *Client) AddCert(cert tls.Certificate) *Client {
	c.judge2genTlsConfig()
	c.transport.TLSClientConfig.Certificates = append(c.transport.TLSClientConfig.Certificates, cert)
	return c
}

func (c *Client) judge2genTlsConfig() {
	if c.transport.TLSClientConfig == nil {
		c.transport.TLSClientConfig = new(tls.Config)
	}
}

func (c *Client) judge2genPool() {
	c.judge2genTlsConfig()
	if c.transport.TLSClientConfig.RootCAs == nil {
		c.transport.TLSClientConfig.RootCAs = x509.NewCertPool()
	}
}
