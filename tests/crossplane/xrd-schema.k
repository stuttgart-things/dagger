__xrd_group = "resources.stuttgart-things.com"
__xrd_kind = "CloudInit"
__xrd_version = "v1alpha1"
__xrd_scope = "Namespaced"
__xrd_plural = "cloudinits"
__xrd_singular = "cloudinit"

"""
CloudInit Composite Resource Definition
Provides cloud-init configuration rendered into a Kubernetes Secret.
"""

# ------------------------------------------------------------
# Spec
# ------------------------------------------------------------

# @xrd
schema CloudInit:
    # Name of the ProviderConfig to use for the Kubernetes provider
    providerConfigRef: str
    # Name of the VM (used for secret naming)
    vmName: str
    # Namespace where the cloud-init Secret will be created
    namespace?: str = "vms"
    # Hostname assigned to the VM
    hostname: str
    # DNS domain appended to the hostname
    domain?: str = "local"
    # System timezone (e.g. UTC, Europe/Berlin)
    timezone?: str = "UTC"
    # User accounts to be created on the system
    users?: [User] = []
    # OS packages to install during provisioning
    packages?: [str] = []
    # Files to be written to disk during provisioning
    writeFiles?: [WriteFile] = []
    # Commands executed late in boot (cloud-init runcmd)
    runcmd?: [str] = []
    # Commands executed early in boot (cloud-init bootcmd)
    bootcmd?: [str] = []
    # Update package cache before installing packages
    packageUpdate?: bool = True
    # Upgrade all installed packages
    packageUpgrade?: bool = False
    # Enable SSH password authentication
    sshPasswordAuth?: bool = False
    # Disable root SSH login
    disableRoot?: bool = True
    # Cloud-init network configuration (YAML string)
    networkConfig?: str = ""

# ------------------------------------------------------------
# Nested Schemas
# ------------------------------------------------------------

schema User:
    # Username to create
    name?: str
    # Authorized SSH public keys for the user
    sshAuthorizedKeys?: [str] = []
    # Sudo configuration (e.g. ALL=(ALL) NOPASSWD:ALL)
    sudo?: str = "ALL=(ALL) NOPASSWD:ALL"
    # Unix groups the user should belong to
    groups?: str = "users"
    # Login shell for the user
    shell?: str = "/bin/bash"
    # User password (hashed)
    password?: str
    # Whether to lock password authentication
    lockPasswd?: bool = False

schema WriteFile:
    # Absolute path of the file to write
    path?: str
    # File content
    content?: str
    # File owner in user:group format
    owner?: str = "root:root"
    # File permissions (octal string)
    permissions?: str = "0644"
    # File content encoding
    encoding?: str = "text/plain"

# ------------------------------------------------------------
# Status
# ------------------------------------------------------------

# @status
schema CloudInitStatus:
    # Name of the generated Kubernetes Secret containing cloud-init data
    secretName?: str = ""
    # Indicates whether the cloud-init Secret is ready
    ready?: bool = False
