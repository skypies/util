A bunch of small helper libraries for Google Cloud Platform APIs.

These should all be usable outside of appengine.

TODO:

* ae.SaveSingletonToMemcacheHandler - Move this stuff out of the way, into DSProvider
* metar - all messed up
* ui/batch, backend/foia - calls appengine/taskqueue
* backend/fr24poller - calls appengine/memcache
* ref/ref.go - calls ae.Singleton stuff

Thoughts on ref.
- ref calls singleton code direct from ae/ds
- there is some singleton code in gcp/ds, that stubs out memcache

1. we could make the high level {Load,Save,Delete}Singleton part of provider interface
It would be hacky - we'd have clones of code in both ae/ds/singleton and gcp/ds/singleton
Could modernize the call sites though

2. We could make just memcache.Put/Get/Delete in the provider interface
That should be simple / small enough
We could even fully implement the HTTP hacky thing in gcp/ds
We could then put the singleton code in util/singleton
We might have a strategy for other memcache usage, too (e.g. metar)
Let's do it

ARGH, this ended up all the way into pi/airspace, which does memcache things.
It will all need to be taught about ds.DatastoreProvider junk.
I just redid (again) the ref/ stuff, to take ds.DatastoreProvider args everywhere, but can't tell if it works yet.

INSTEAD - put an explicit shium into util/ae, that trampolines into util/singleton with a 'p'

Sigh, we need {Get,Set}Multi routines too.

How about we just delete all the memcache, everywhere ?
- need it for the realtime airspace. But we can do that with the
hardhacked URL thing we have today ... and the singletonhandler thing.

Too tired to make a decision. It's all annoying crap messy code.
