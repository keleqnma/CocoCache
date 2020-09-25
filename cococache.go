package cococache

import (
	"fmt"
	"log"
	"sync"
)

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type CocoCache struct {
	//callback function
	getter Getter
	//cache
	mainCache cache
	peers     PeerPicker
}

var (
	mu        sync.RWMutex
	cocoCache *CocoCache
)

// NewCache create a new instance of Group
func NewCache(cacheBytes int64, getter Getter) *CocoCache {
	mu.Lock()
	defer mu.Unlock()
	cocoCache := &CocoCache{
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	return cocoCache
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *CocoCache) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// GetCache returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetCache() *CocoCache {
	return cocoCache
}

func (g *CocoCache) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}

	return g.load(key)
}

func (g *CocoCache) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[CocoCache] Failed to get from peer", err)
		}
	}
	return g.getLocally(key)
}

func (g *CocoCache) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: bytes}
	g.populateCache(key, value)
	return value, nil
}

func (g *CocoCache) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *CocoCache) populateCache(key string, value ByteView) {
	g.mainCache.set(key, value)
}
