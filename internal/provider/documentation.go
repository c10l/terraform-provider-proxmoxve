package provider

const providerMarkdownDescription = "" +
	"Use the ProxmoxVE provider to manage Proxmox Virtual Environment configuration, reasources, virtual machines etc. on a compatible PMVE cluster or standalone server. You must configure the provider with API credentials before using it." +
	"<p />The following environment variables can be set as a fallback for any omitted attributes in the provider declaration: `PROXMOXVE_BASE_URL`, `PROXMOXVE_TOKEN_ID`, `PROXMOXVE_SECRET`, `PROXMOXVE_ROOT_PASSWORD`, `PROXMOXVE_TOTPSEED`, `PROXMOXVE_TLS_INSECURE`." +
	"<p />**NOTE:** `base_url` attribute is always required. Additionally, most API endpoints require `token_id` and `secret`. Other API endpoints require `root_password`, and if 2FA is enabled for the `root` user, `totp_seed` must also be informed."

const docRequiresRoot = "<p />**NOTE:** This resource requires the provider attribute `root_password` or the environment variable `PROXMOXVE_ROOT_PASSWORD` set."
