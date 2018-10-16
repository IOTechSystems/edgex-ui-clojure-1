# EdgeX Manager UI

The main project source is in `src/main`. Built using the Fulcro Clojure framework
[http://fulcro.fulcrologic.com/](http://fulcro.fulcrologic.com/). The Fulcro web site has a Developer's Guide
and links to a series of Youtube training videos.

The Clojure build tool [Leiningen](https://leiningen.org/) is used.

A brief introduction to the GUI, including using a pre-built docker image, is available [here](docs/EdgeXpertManager.rst).

The UI was developed using CSS from the commercial version of [Creative Tim](https://www.creative-tim.com/)'s Light
Bootstrap Dashboard for Bootstrap 3. Note that the commercial version is now for Bootstrap 4.
The CSS from the free open source version of the
[dashboard](https://github.com/creativetimofficial/light-bootstrap-dashboard) has been substituted and a collapsible
menu feature disabled.

```
├── config                     server configuration files. See Fulcro docs.
│   ├── defaults.edn
│   ├── dev.edn                development endpoint configuration (localhost)
│   └── prod.edn               production endpoint configuration (docker host names)
└── manager
    ├── api
    │   ├── edgex.clj          EdgeX REST API functions
    │   ├── file_db.clj        file upload support
    │   ├── mutations.clj      server-side version of mutations
    │   ├── mutations.cljs     client-side version of mutations
    │   ├── read.clj           server implementation of reads
    │   └── util.cljc          common utility functions
    ├── client.cljs            client creation (shared among dev/prod)
    ├── client_main.cljs       production client main
    ├── server.clj             server creation (shared among dev/prod)
    ├── server_main.clj        production server main
    └── ui
        ├── addressables.cljs      Addressable table and editing
        ├── commands.cljs          Command support
        ├── common.cljs            Common ui functions and constants
        ├── date_time_picker.cljs  Date and Time picker component
        ├── devices.cljs           Device table
        ├── dialogs.cljs           Common modal dialog functions
        ├── endpoints.cljs         Service endpoint editing
        ├── exports.cljs           Exports table and editing
        ├── graph.cljs             Graphing functions
        ├── ident.cljc             Identity function
        ├── labels.cljs            Form labeling
        ├── logging.cljs           Log entry display
        ├── main.cljs              Main page
        ├── notifications.cljs     Notification display
        ├── profiles.cljs          Profile tables and editing
        ├── readings.cljs          Readings table
        ├── root.cljs              UI root component
        ├── routing.cljs           HTML routing support
        ├── schedules.cljs         Schedule tables and editing
        └── table.cljc             Table macro and utility functions
```

## Development Mode

Special code for working in dev mode is in `src/dev`, which is not on
the build for production builds.

Running all client builds:

```
JVM_OPTS="-Ddev -Dtest -Dcards" lein run -m clojure.main script/figwheel.clj
dev:cljs.user=> (log-app-state) ; show current state of app
dev:cljs.user=> :cljs/quit      ; switch REPL to a different build
```

Or use a plain REPL in IntelliJ with JVM options of `-Ddev -Dtest -Dcards` and parameters of
`script/figwheel.clj`.

In emacs run cider-jack-in on src/main/manager/client.cljs. Start figwheel with:
```
user> (start-figwheel)
```

For a faster hot code reload experience, run only the build that matters to you at the time,

Running multiple builds in one figwheel can slow down hot code reload. You can also
run multiple separate figwheel instances to leverage more of your CPU cores, and
an additional system property can be used to allow this (by allocating different network ports
to figwheel instances):

```
# Assuming one per terminal window...each gets a REPL that expects STDIN/STDOUT.
JVM_OPTS="-Ddev -Dfigwheel.port=8081" lein run -m clojure.main script/figwheel.clj
JVM_OPTS="-Dtest -Dfigwheel.port=8082" lein run -m clojure.main script/figwheel.clj
JVM_OPTS="-Dcards -Dfigwheel.port=8083" lein run -m clojure.main script/figwheel.clj
```

Running the server:

Start a clj REPL in IntelliJ, or from the command line:

```
lein run -m clojure.main
user=> (go)
...
user=> (restart) ; stop, reload server code, and go again
user=> (tools-ns/refresh) ; retry code reload if hot server reload fails
```

In emacs run cider-jack-in on src/main/manager/server.clj.
```
user> (go)
```

The URLs are:

- Client (using server): [http://localhost:3000](http://localhost:3000)
- Cards: [http://localhost:3449/cards.html](http://localhost:3449/cards.html)

## Dev Cards

Dev cards can be used as a standalone development playground. See [https://github.com/bhauman/devcards](https://github.com/bhauman/devcards).
Dev cards are located in `src/cards`.

```
JVM_OPTS="-Dcards" lein run -m clojure.main script/figwheel.clj
```

Or use a plain REPL in IntelliJ with JVM options of `-Dcards` and parameters of
`script/figwheel.clj`.

To add a new card namespace, remember to add a require for it to the `cards.cljs` file.

## Standalone Runnable Jar (Production, with advanced optimized client js)

```
lein uberjar
java -jar target/edgex_manager.jar
```
