```
├───.assets              — moonpool project media assets (logo, exe icon, etc)
├───.build               — scripts related to building moonpool
├───.docs                — documentation
├───.github              — github related files (actions, workflows, etc)
├───.internal-dev        — .gitignore'd folder to store personal notes locally.  wont be published to main repo
├───.scripts             - scripts related to externally testing moonpool (won't ever be  invoked with 'go test')
├───.vscode              — config files for visual studio code to aid in development
├───api                  — public facing package to interface with moonpool. all non-database related operations should occur here if possible.
├───cmd                  — cli for moonpool
│   └───moonpool         — main.go resides here
├───config               — json configuration for moonpool. contains default values
├───entry                - public facing package to "tie" all common structs into a unified package. mostly used by api package but sometimes in db for performance reasons 
├───importer             - struct that satisfies the api package "importer" interface requirements. 
└───internal             - low-level packages that contain the main functionality of moonpool
    ├───db               - database layer that contains the "heart" of moonpool
    │   ├───archive      - handles storing and retriving everything relating to tags, timestamps, hashes, etc. it is where each 'entry' or 'archive' gets stored into. 
    │   └───thumbnail    - handles storing and retriving thumbnails
    ├───file             - everything related to files in terms of timestamps, hashing, copying, etc
    ├───log              - verything relating to logging
    ├───media            - handles everything relating to image manipulation
    │   └───thumbnail    - only contains a struct for thumbnails. might be deprecated in the future
    ├───profile          - performance profiling & measurement
    ├───server           - contains webUI and webAPI services
    └───www              - webUI (what the user sees)
        ├───assets       - static files to be served to user
        │   ├───scripts  - javascript files (currently unused)
        │   └───static   - css files and a 404.png placeholder for missing thumbnails/media
        └───templates    - golang html template files
```