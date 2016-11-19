// Package routecache is meant to assist in resolving the most used routes at
// an application. Implemented as a LRU, it returns always its full context for
// iteration at the router handler.
package routecache

// based on groupcache's LRU

import (
	"container/list"

	"github.com/iron-io/functions/api/models"
)

// Cache holds an internal linkedlist for hotness management. It is not safe
// for concurrent use, must be guarded externally.
type Cache struct {
	MaxEntries int

	ll     *list.List
	cache  map[string]*list.Element
	values []*models.Route
}

type routecacheentry struct {
	route *models.Route
}

// New returns a route cache.
func New(maxentries int) *Cache {
	return &Cache{
		MaxEntries: maxentries,
		ll:         list.New(),
		cache:      make(map[string]*list.Element),
	}
}

// Refresh updates internal linkedlist either adding a new route to the front,
// or moving it to the front when used. It will discard seldom used routes.
func (c *Cache) Refresh(route *models.Route) {
	if c.cache == nil {
		return
	}

	if ee, ok := c.cache[route.Path]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*routecacheentry).route = route
		c.updatevalues()
		return
	}

	ele := c.ll.PushFront(&routecacheentry{route})
	c.cache[route.Path] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.removeOldest()
	}

	c.updatevalues()
}

// Get looks up a path's route from the cache.
func (c *Cache) Get(path string) (route *models.Route, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[path]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*routecacheentry).route, true
	}
	return
}

func (c *Cache) updatevalues() {
	c.values = make([]*models.Route, 0, c.ll.Len())
	for e := c.ll.Front(); e != nil; e = e.Next() {
		route := e.Value.(*routecacheentry).route
		c.values = append(c.values, route)
	}
}

func (c *Cache) removeOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*routecacheentry)
	delete(c.cache, kv.route.Path)
}
