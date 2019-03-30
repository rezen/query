package query

import (
  "net/http"
  "strings"
)

var cdnDomains = []CdnDomain{
  {".akamai.net", "Akamai"},
  {".akamaized.net", "Akamai"},
  {".akamaiedge.net", "Akamai"},
  {".akamaihd.net", "Akamai"},
  {".edgesuite.net", "Akamai"},
  {".edgekey.net", "Akamai"},
  {".srip.net", "Akamai"},
  {".akamaitechnologies.com", "Akamai"},
  {".akamaitechnologies.fr", "Akamai"},
  {".tl88.net", "Akamai China CDN"},
  {".llnwd.net", "Limelight"},
  {"edgecastcdn.net", "Edgecast"},
  {".systemcdn.net", "Edgecast"},
  {".transactcdn.net", "Edgecast"},
  {".v1cdn.net", "Edgecast"},
  {".v2cdn.net", "Edgecast"},
  {".v3cdn.net", "Edgecast"},
  {".v4cdn.net", "Edgecast"},
  {".v5cdn.net", "Edgecast"},
  {"hwcdn.net", "Highwinds"},
  {".simplecdn.net", "Simple CDN"},
  {".instacontent.net", "Mirror Image"},
  {".footprint.net", "Level 3"},
  {".fpbns.net", "Level 3"},
  {".ay1.b.yahoo.com", "Yahoo"},
  {".yimg.", "Yahoo"},
  {".yahooapis.com", "Yahoo"},
  {".google.", "Google"},
  {"googlesyndication.", "Google"},
  {"youtube.", "Google"},
  {".googleusercontent.com", "Google"},
  {"googlehosted.com", "Google"},
  {".gstatic.com", "Google"},
  {".doubleclick.net", "Google"},
  {".insnw.net", "Instart Logic"},
  {".inscname.net", "Instart Logic"},
  {".internapcdn.net", "Internap"},
  {".cloudfront.net", "Amazon CloudFront"},
  {".netdna-cdn.com", "NetDNA"},
  {".netdna-ssl.com", "NetDNA"},
  {".netdna.com", "NetDNA"},
  {".kxcdn.com", "KeyCDN"},
  {".cotcdn.net", "Cotendo CDN"},
  {".cachefly.net", "Cachefly"},
  {"bo.lt", "BO.LT"},
  {".cloudflare.com", "Cloudflare"},
  {".afxcdn.net", "afxcdn.net"},
  {".lxdns.com", "ChinaNetCenter"},
  {".wscdns.com", "ChinaNetCenter"},
  {".wscloudcdn.com", "ChinaNetCenter"},
  {".ourwebpic.com", "ChinaNetCenter"},
  {".att-dsa.net", "AT&T"},
  {".vo.msecnd.net", "Microsoft Azure"},
  {".azureedge.net", "Microsoft Azure"},
  {".azure.microsoft.com", "Microsoft Azure"},
  {".voxcdn.net", "VoxCDN"},
  {".bluehatnetwork.com", "Blue Hat Network"},
  {".swiftcdn1.com", "SwiftCDN"},
  {".swiftserve.com", "SwiftCDN"},
  {".cdngc.net", "CDNetworks"},
  {".gccdn.net", "CDNetworks"},
  {".panthercdn.com", "CDNetworks"},
  {".fastly.net", "Fastly"},
  {".fastlylb.net", "Fastly"},
  {".nocookie.net", "Fastly"},
  {".gslb.taobao.com", "Taobao"},
  {".gslb.tbcache.com", "Alimama"},
  {".mirror-image.net", "Mirror Image"},
  {".yottaa.net", "Yottaa"},
  {".cubecdn.net", "cubeCDN"},
  {".cdn77.net", "CDN77"},
  {".cdn77.org", "CDN77"},
  {".incapdns.net", "Incapsula"},
  {".bitgravity.com", "BitGravity"},
  {".r.worldcdn.net", "OnApp"},
  {".r.worldssl.net", "OnApp"},
  {"tbcdn.cn", "Taobao"},
  {".taobaocdn.com", "Taobao"},
  {".ngenix.net", "NGENIX"},
  {".pagerain.net", "PageRain"},
  {".ccgslb.com", "ChinaCache"},
  {"cdn.sfr.net", "SFR"},
  {".azioncdn.net", "Azion"},
  {".azioncdn.com", "Azion"},
  {".azion.net", "Azion"},
  {".cdncloud.net.au", "MediaCloud"},
  {".rncdn1.com", "Reflected Networks"},
  {".rncdn7.com", "Reflected Networks"},
  {".cdnsun.net", "CDNsun"},
  {".mncdn.com", "Medianova"},
  {".mncdn.net", "Medianova"},
  {".mncdn.org", "Medianova"},
  {"cdn.jsdelivr.net", "jsDelivr"},
  {".nyiftw.net", "NYI FTW"},
  {".nyiftw.com", "NYI FTW"},
  {".resrc.it", "ReSRC.it"},
  {".zenedge.net", "Zenedge"},
  {".lswcdn.net", "LeaseWeb CDN"},
  {".lswcdn.eu", "LeaseWeb CDN"},
  {".revcn.net", "Rev Software"},
  {".revdn.net", "Rev Software"},
  {".caspowa.com", "Caspowa"},
  {".twimg.com", "Twitter"},
  {".facebook.com", "Facebook"},
  {".facebook.net", "Facebook"},
  {".fbcdn.net", "Facebook"},
  {".cdninstagram.com", "Facebook"},
  {".rlcdn.com", "Reapleaf"},
  {".wp.com", "WordPress"},
  {".wordpress.com", "WordPress"},
  {".gravatar.com", "WordPress"},
  {".aads1.net", "Aryaka"},
  {".aads-cn.net", "Aryaka"},
  {".aads-cng.net", "Aryaka"},
  {".squixa.net", "section.io"},
  {".bisongrid.net", "Bison Grid"},
  {".cdn.gocache.net", "GoCache"},
  {".hiberniacdn.com", "HiberniaCDN"},
  {".cdntel.net", "Telenor"},
  {".raxcdn.com", "Rackspace"},
  {".unicorncdn.net", "UnicornCDN"},
  {".optimalcdn.com", "Optimal CDN"},
  {".kinxcdn.com", "KINX CDN"},
  {".kinxcdn.net", "KINX CDN"},
  {".stackpathdns.com", "StackPath"},
  {".hosting4cdn.com", "Hosting4CDN"},
  {".netlify.com", "Netlify"},
  {".b-cdn.net", "BunnyCDN"},
  {".pix-cdn.org", "Advanced Hosters CDN"},
  {".roast.io", "Roast.io"},
  {".cdnvideo.ru", "CDNvideo"},
  {".cdnvideo.net", "CDNvideo"},
  {".trbcdn.ru", "TRBCDN"},
  {".cedexis.net", "Cedexis"},
  {".streamprovider.net", "Rocket CDN"},
  {".singularcdn.net.br", "Singular CDN"},
  {".googleapis.", "Google"},
}

var cdnHeaders = []CdnHeader{
  {"server", "cloudflare", "Cloudflare"},
  {"server", "yunjiasu", "Yunjiasu"},
  {"server", "ECS", "Edgecast"},
  {"server", "ECAcc", "Edgecast"},
  {"server", "ECD", "Edgecast"},
  {"server", "NetDNA", "NetDNA"},
  {"server", "Airee", "Airee"},
  {"X-CDN-Geo", "", "OVH CDN"},
  {"X-CDN-Pop", "", "OVH CDN"},
  {"X-Px", "", "CDNetworks"},
  {"X-Instart-Request-ID", "instart", "Instart Logic"},
  {"Via", "CloudFront", "Amazon CloudFront"},
  {"X-Edge-IP", "", "CDN"},
  {"X-Edge-Location", "", "CDN"},
  {"X-HW", "", "Highwinds"},
  {"X-Powered-By", "NYI FTW", "NYI FTW"},
  {"X-Delivered-By", "NYI FTW", "NYI FTW"},
  {"server", "ReSRC", "ReSRC.it"},
  {"X-Cdn", "Zenedge", "Zenedge"},
  {"server", "leasewebcdn", "LeaseWeb CDN"},
  {"Via", "Rev-Cache", "Rev Software"},
  {"X-Rev-Cache", "", "Rev Software"},
  {"Server", "Caspowa", "Caspowa"},
  {"Server", "SurgeCDN", "Surge"},
  {"server", "sffe", "Google"},
  {"server", "gws", "Google"},
  {"server", "GSE", "Google"},
  {"server", "Golfe2", "Google"},
  {"Via", "google", "Google"},
  {"server", "tsa_b", "Twitter"},
  {"X-Cache", "cache.51cdn.com", "ChinaNetCenter"},
  {"X-CDN", "Incapsula", "Incapsula"},
  {"X-Iinfo", "", "Incapsula"},
  {"X-Ar-Debug", "", "Aryaka"},
  {"server", "gocache", "GoCache"},
  {"server", "hiberniacdn", "HiberniaCDN"},
  {"server", "UnicornCDN", "UnicornCDN"},
  {"server", "Optimal CDN", "Optimal CDN"},
  {"server", "Sucuri/Cloudproxy", "Sucuri Firewall"},
  {"x-sucuri-id", "", "Sucuri Firewall"},
  {"server", "Netlify", "Netlify"},
  {"section-io-id", "", "section.io"},
  {"server", "Testa/", "Naver"},
  {"server", "BunnyCDN", "BunnyCDN"},
  {"server", "MNCDN", "Medianova"},
  {"server", "Roast.io", "Roast.io"},
  {"server", "SingularCDN", "Singular CDN"},
  {"x-rocket-node", "", "Rocket CDN"},
}

// https://gist.github.com/fffaraz/a12e74093fd19f773dd7
// https://github.com/WPO-Foundation/webpagetest/blob/master/agent/wpthook/cdn.h
type CdnDomain struct {
  Domain string
  Name   string
}

func (c CdnDomain) IsMatch(check string) bool {
  // @todo improve
  return strings.Contains(check, c.Domain)
}

type CdnHeader struct {
  Header string
  Value  string
  Name   string
}

func (c CdnHeader) IsMatch(r http.Response) bool {
  return r.Header.Get(c.Header) == c.Value
}

func IsCdnDomain(check string) (bool, string) {
  for _, c := range cdnDomains {
    if c.IsMatch(check) {
      return true, c.Name
    }
  }

  return false, ""
}

func IsResponseFromCdn(r http.Response) (bool, string) {
  for _, c := range cdnHeaders {
    if c.IsMatch(r) {
      return true, c.Name
    }
  }

  return false, ""
}
