
export default class Tendbha {
  bk_biz_id: number;
  bk_biz_name: string;
  bk_cloud_id: number;
  bk_cloud_name: string;
  cluster_name: string;
  cluster_type: string;
  create_at: string;
  creator: string;
  db_module_name: string;
  db_module_id: number;
  id: number;
  master_domain: string;
  masters: {
    bk_biz_id: number,
    bk_cloud_id: number,
    bk_host_id: number,
    bk_instance_id: number,
    instance: string,
    ip: string,
    name: string,
    phase: string,
    port: number,
    spec_config: Record<'id', number>,
    status: string,
  }[];
  operations: Array<{
    cluster_id: number,
    flow_id: number,
    operator: string,
    status: string,
    ticket_id: number,
    ticket_type: string,
    title: string,
  }>;
  permission: Record<'mysql_authorize' | 'mysql_destroy' | 'mysql_enable_disable' | 'mysql_view', boolean>;
  phase: string;
  phase_name: string;
  proxies: Tendbha['masters'];
  slave_domain: string;
  slaves: Tendbha['masters'];
  status: string;

  constructor(payload = {} as Tendbha) {
    this.bk_biz_id = payload.bk_biz_id || 0;
    this.bk_biz_name = payload.bk_biz_name || '';
    this.bk_cloud_id = payload.bk_cloud_id || 0;
    this.bk_cloud_name = payload.bk_cloud_name || '';
    this.cluster_name = payload.cluster_name || '';
    this.cluster_type = payload.cluster_type || '';
    this.create_at = payload.create_at || '';
    this.creator = payload.creator || '';
    this.db_module_name = payload.db_module_name || '';
    this.db_module_id = payload.db_module_id || 0;
    this.id = payload.id || 0;
    this.master_domain = payload.master_domain || '';
    this.masters = payload.masters || [];
    this.operations = payload.operations || [];
    this.permission = payload.permission || {};
    this.phase = payload.phase || '';
    this.phase_name = payload.phase_name || '';
    this.proxies = payload.proxies || [];
    this.slave_domain = payload.slave_domain || '';
    this.slaves = payload.slaves || [];
    this.status = payload.status || '';
  }

  get isOnline() {
    return this.phase === 'online';
  }

  get masterDomainDisplayName() {
    const port = this.proxies[0]?.port;
    const displayName = port ? `${this.master_domain}:${port}` : this.master_domain;
    return displayName;
  }

  get slaveDomainDisplayName() {
    const port = this.proxies[0]?.port;
    const displayName = port ? `${this.slave_domain}:${port}` : this.slave_domain;
    return displayName;
  }
}
