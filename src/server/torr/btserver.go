package torr

import (
	"log"
	"sync"

	"server/settings"
	"server/torr/storage/torrstor"
	"server/torr/utils"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type BTServer struct {
	config *torrent.ClientConfig
	client *torrent.Client

	storage *torrstor.Storage

	torrents map[metainfo.Hash]*Torrent

	mu sync.Mutex
}

func NewBTS() *BTServer {
	bts := new(BTServer)
	bts.torrents = make(map[metainfo.Hash]*Torrent)
	return bts
}

func (bt *BTServer) Connect() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	var err error
	bt.configure()
	bt.client, err = torrent.NewClient(bt.config)
	bt.torrents = make(map[metainfo.Hash]*Torrent)
	return err
}

func (bt *BTServer) Disconnect() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.client != nil {
		bt.client.Close()
		bt.client = nil
		utils.FreeOSMemGC()
	}
}

func (bt *BTServer) Reconnect() error {
	bt.Disconnect()
	return bt.Connect()
}

func (bt *BTServer) configure() {
	bt.storage = torrstor.NewStorage(settings.BTsets.CacheSize)

	blocklist, _ := utils.ReadBlockedIP()

	userAgent := "uTorrent/3.5.5"
	peerID := "-UT3550-"
	cliVers := "µTorrent 3.5.5"

	bt.config = torrent.NewDefaultClientConfig()

	bt.config.Debug = settings.BTsets.EnableDebug
	bt.config.DisableIPv6 = settings.BTsets.EnableIPv6 == false
	bt.config.DisableTCP = settings.BTsets.DisableTCP
	bt.config.DisableUTP = settings.BTsets.DisableUTP
	bt.config.NoDefaultPortForwarding = settings.BTsets.DisableUPNP
	bt.config.NoDHT = settings.BTsets.DisableDHT
	bt.config.NoUpload = settings.BTsets.DisableUpload
	// bt.config.EncryptionPolicy = torrent.EncryptionPolicy{
	// 	DisableEncryption: settings.BTsets.Encryption == 1,
	// 	ForceEncryption:   settings.BTsets.Encryption == 2,
	// }
	bt.config.IPBlocklist = blocklist
	bt.config.DefaultStorage = bt.storage
	bt.config.Bep20 = peerID
	bt.config.PeerID = utils.PeerIDRandom(peerID)
	bt.config.HTTPUserAgent = userAgent
	bt.config.ExtendedHandshakeClientVersion = cliVers
	bt.config.EstablishedConnsPerTorrent = settings.BTsets.ConnectionsLimit
	if settings.BTsets.DhtConnectionLimit > 0 {
		bt.config.ConnTracker.SetMaxEntries(settings.BTsets.DhtConnectionLimit)
	}
	if settings.BTsets.DownloadRateLimit > 0 {
		bt.config.DownloadRateLimiter = utils.Limit(settings.BTsets.DownloadRateLimit * 1024)
	}
	if settings.BTsets.UploadRateLimit > 0 {
		bt.config.UploadRateLimiter = utils.Limit(settings.BTsets.UploadRateLimit * 1024)
	}
	if settings.BTsets.PeersListenPort > 0 {
		bt.config.ListenPort = settings.BTsets.PeersListenPort
	}

	log.Println("Configure client:", settings.BTsets)
}

func (bt *BTServer) RemoveTorrent(hash torrent.InfoHash) {
	if torr, ok := bt.torrents[hash]; ok {
		torr.Close()
	}
}
