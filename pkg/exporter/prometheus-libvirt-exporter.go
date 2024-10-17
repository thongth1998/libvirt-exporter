package exporter

import (
	"encoding/xml"
	"regexp"
	"time"
	"strconv"
	"strings"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/inovex/prometheus-libvirt-exporter/libvirt_schema"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "libvirt_domain"

var (
	libvirtUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("", "", "up"),
		"Whether scraping libvirt's metrics was successful.",
		nil,
		nil)

	libvirtDomainNumbers = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "", "domains"),
		"Number of domains",
		nil,
		nil)
	libvirtDomainBlockRdTotalTimeSecondsDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "read_time_seconds_total"),
                "Total time spent on reads from a block device, in seconds.",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
	libvirtDomainBlockWrTotalTimeSecondsDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "write_time_seconds_total"),
                "Total time spent on writes on a block device, in seconds",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
	libvirtDomainMemoryStatDiskCachesBytesDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "memory_stats", "disk_cache_bytes"),
                "The amount of memory, that can be quickly reclaimed without additional I/O (in bytes)."+
                        "Typically these pages are used for caching files from disk.",
                []string{"domain", "instance_name", "project_id", "project_name"},
                nil)
        libvirtDomainMemoryStatUsedPercentDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "memory_stats", "used_percent"),
                "The amount of memory in percent, that used by domain.",
                []string{"domain", "instance_name", "project_id", "project_name"},
                nil)

	//domain info
	libvirtDomainState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "info", "state"),
		"Code of the domain state",
		[]string{"domain", "instance_name", "project_id", "project_name", "state_desc"},
		nil)
	libvirtDomainInfoMaxMemDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "info", "maximum_memory_bytes"),
		"Maximum allowed memory of the domain, in bytes.",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainInfoMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "info", "memory_usage_bytes"),
		"Memory usage of the domain, in bytes.",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainInfoNrVirtCpuDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "info", "virtual_cpus"),
		"Number of virtual CPUs for the domain.",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainInfoCpuTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "info", "cpu_time_seconds_total"),
		"Amount of CPU time used by the domain, in seconds.",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)

	//domain memory stats
	libvirtDomainMemoryStatsSwapInBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "memory_stats", "swap_in_bytes"),
		"Memory swapped in for this domain(the total amount of data read from swap space)",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainMemoryStatsSwapOutBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "memory_stats", "swap_out_bytes"),
		"Memory swapped out for this domain (the total amount of memory written out to swap space)",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainMemoryStatsUnusedBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "memory_stats", "unused_bytes"),
		"Memory unused by the domain",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainMemoryStatsAvailableInBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "memory_stats", "available_bytes"),
		"Memory available to the domain",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainMemoryStatsUsableBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "memory_stats", "usable_bytes"),
		"Memory usable by the domain (corresponds to 'Available' in /proc/meminfo)",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainMemoryStatsRssBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "memory_stats", "rss_bytes"),
		"Resident Set Size of the process running the domain",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
        libvirtDomainMemoryStatFreePercentDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "memory_stats", "free_percent"),
                "The percentage of memory currently available for use by the instance",
                []string{"domain", "instance_name", "project_id", "project_name"},
                nil)
	libvirtDomainMemoryStatUsednocachePercentDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "memory_stats", "usednocache_percent"),
                "The percentage of memory currently used without pagecache/buffercache by the instance",
                []string{"domain", "instance_name", "project_id", "project_name"},
                nil)

	//domain block stats
	libvirtDomainBlockStatsInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "block_stats", "info"),
		"Metadata information on block devices.",
		[]string{"domain", "instance_name", "project_id", "project_name", "disk_type", "target_bus", "driver_name", "driver_type", "driver_cache", "driver_discard", "source_file", "source_protocol", "target_device", "serial"},
		nil)
	libvirtDomainBlockStatsRdBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "block_stats", "read_bytes_total"),
		"Number of bytes read from a block device, in bytes.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device"},
		nil)
	libvirtDomainBlockStatsRdReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "block_stats", "read_requests_total"),
		"Number of read requests from a block device.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device"},
		nil)
	libvirtDomainBlockStatsWrBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "block_stats", "write_bytes_total"),
		"Number of bytes written from a block device, in bytes.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device"},
		nil)
	libvirtDomainBlockStatsWrReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "block_stats", "write_requests_total"),
		"Number of write requests from a block device.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device"},
		nil)
        libvirtDomainBlockCapacityBytesDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "capacity_bytes"),
                "Logical size in bytes of the block device",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
        libvirtDomainBlockTotalBytesSecDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "limit_total_bytes"),
                "Total throughput limit in bytes per second",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
        libvirtDomainBlockWriteBytesSecDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "limit_write_bytes"),
                "Write throughput limit in bytes per second",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
        libvirtDomainBlockReadBytesSecDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "limit_read_bytes"),
                "Read throughput limit in bytes per second",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
        libvirtDomainBlockTotalIopsSecDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "limit_total_requests"),
                "Total requests per second limit",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
        libvirtDomainBlockWriteIopsSecDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "limit_write_requests"),
                "Write requests per second limit",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
        libvirtDomainBlockReadIopsSecDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "limit_read_requests"),
                "Read requests per second limit",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
	libvirtDomainBlockReadBytesPercentDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "read_bytes_usage_percent"),
                "The percentage of read bytes usage to the read throughput limit.",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
	libvirtDomainBlockWriteBytesPercentDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "write_bytes_usage_percent"),
                "The percentage of write bytes usage to the write throughput limit.",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
	libvirtDomainBlockTotalBytesPercentDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "block_stats", "total_bytes_usage_percent"),
                "The percentage of total bytes usage to the total throughput limit.",
                []string{"domain", "instance_name", "project_id", "project_name", "target_device"},
                nil)
	libvirtDomainBlockReadRequestsPercentDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "block_stats", "read_requests_usage_percent"),
		"The percentage of read requests usage to the read IOPS limit.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device"},
		nil)
	libvirtDomainBlockWriteRequestsPercentDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "block_stats", "write_requests_usage_percent"),
		"The percentage of write requests usage to the IOPS limit.",
			[]string{"domain", "instance_name", "project_id", "project_name", "target_device"},
		nil)

	libvirtDomainBlockTotalRequestsPercentDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "block_stats", "total_requests_usage_percent"),
		"The percentage of total requests usage to the total IOPS limit.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device"},
		nil)


	//domain interface stats
	libvirtDomainInterfaceInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "interface_stats", "info"),
		"Metadata on network interfaces.",
		[]string{"domain", "instance_name", "project_id", "project_name", "interface_type", "source_bridge", "target_device", "mac_address", "model_type", "mtu_size", "alias_name"},
		nil)
	libvirtDomainInterfaceRxBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "interface_stats", "receive_bytes_total"),
		"Number of bytes received on a network interface, in bytes.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device", "alias_name"},
		nil)
	libvirtDomainInterfaceRxPacketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "interface_stats", "receive_packets_total"),
		"Number of packets received on a network interface.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device", "alias_name"},
		nil)
	libvirtDomainInterfaceRxErrsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "interface_stats", "receive_errors_total"),
		"Number of packet receive errors on a network interface.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device", "alias_name"},
		nil)
	libvirtDomainInterfaceRxDropDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "interface_stats", "receive_drops_total"),
		"Number of packet receive drops on a network interface.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device", "alias_name"},
		nil)
	libvirtDomainInterfaceTxBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "interface_stats", "transmit_bytes_total"),
		"Number of bytes transmitted on a network interface, in bytes.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device", "alias_name"},
		nil)
	libvirtDomainInterfaceTxPacketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "interface_stats", "transmit_packets_total"),
		"Number of packets transmitted on a network interface.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device", "alias_name"},
		nil)
	libvirtDomainInterfaceTxErrsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "interface_stats", "transmit_errors_total"),
		"Number of packet transmit errors on a network interface.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device", "alias_name"},
		nil)
	libvirtDomainInterfaceTxDropDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "interface_stats", "transmit_drops_total"),
		"Number of packet transmit drops on a network interface.",
		[]string{"domain", "instance_name", "project_id", "project_name", "target_device", "alias_name"},
		nil)

	// domain vcpu stats
	libvirtDomainVCPUStatsCurrent = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "vcpu", "current"),
		"Number of current online vCPUs.",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainVCPUStatsMaximum = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "vcpu", "maximum"),
		"Number of maximum online vCPUs.",
		[]string{"domain", "instance_name", "project_id", "project_name"},
		nil)
	libvirtDomainVCPUStatsState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "vcpu", "state"),
		"State of the vCPU.",
		[]string{"domain", "instance_name", "project_id", "project_name", "vcpu"},
		nil)
	libvirtDomainVCPUStatsTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "vcpu", "time_seconds_total"),
		"Time spent by the virtual CPU.",
		[]string{"domain", "instance_name", "project_id", "project_name", "vcpu"},
		nil)
	libvirtDomainVCPUStatsWait = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "vcpu", "wait_seconds_total"),
		"Time the vCPU wants to run, but the host scheduler has something else running ahead of it.",
		[]string{"domain", "instance_name", "project_id", "project_name", "vcpu"},
		nil)
	libvirtDomainVCPUStatsDelay = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "vcpu", "delay_seconds_total"),
		"Time the vCPU spent waiting in the queue instead of running. Exposed to the VM as steal time.",
		[]string{"domain", "instance_name", "project_id", "project_name", "vcpu"},
		nil)
	libvirtDomainVCPUStatsSysPercent = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "vcpu", "sys_percent"),
                "CPU usage percent by instance on all vcpus",
                []string{"domain", "instance_name", "project_id", "project_name", "vcpu"},
                nil)
	libvirtDomainVCPUStatsStealPercent = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, "vcpu", "steal_percent"),
                "The percentage of time the virtual machine process is waiting on the physical CPU for its CPU time",
                []string{"domain", "instance_name", "project_id", "project_name", "vcpu"},
                nil)


	// storage pool stats
	libvirtStoragePoolState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "storage_pool", "state"),
		"State of the storage pool.",
		[]string{"storage_pool"},
		nil)
	libvirtStoragePoolCapacity = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "storage_pool", "capacity_bytes"),
		"Size of the storage pool in logical bytes.",
		[]string{"storage_pool"},
		nil)
	libvirtStoragePoolAllocation = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "storage_pool", "allocation_bytes"),
		"Current allocation bytes of the storage pool.",
		[]string{"storage_pool"},
		nil)
	libvirtStoragePoolAvailable = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "storage_pool", "available_bytes"),
		"Remaining free space of the storage pool in bytes.",
		[]string{"storage_pool"},
		nil)

	// info metrics
	libvirtDomainInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "domain",  "info"),
		"Metadata labels for the domain.",
		[]string{"domain", "instance_name", "project_id", "project_name", "os_type", "os_type_arch", "os_type_machine"},
		nil)

	// info metrics from metadata extracted OpenStack Nova
	libvirtDomainOpenstackInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "openstack_info"),
		"OpenStack Metadata labels for the domain.",
		[]string{"domain",  "instance_name", "instance_id", "flavor_name", "user_name", "user_id", "project_name", "project_id"},
		nil)

	domainState = map[libvirt_schema.DomainState]string{
		libvirt_schema.DOMAIN_NOSTATE:     "no state",
		libvirt_schema.DOMAIN_RUNNING:     "the domain is running",
		libvirt_schema.DOMAIN_BLOCKED:     "the domain is blocked on resource",
		libvirt_schema.DOMAIN_PAUSED:      "the domain is paused by user",
		libvirt_schema.DOMAIN_SHUTDOWN:    "the domain is being shut down",
		libvirt_schema.DOMAIN_SHUTOFF:     "the domain is shut off",
		libvirt_schema.DOMAIN_CRASHED:     "the domain is crashed",
		libvirt_schema.DOMAIN_PMSUSPENDED: "the domain is suspended by guest power management",
		libvirt_schema.DOMAIN_LAST:        "this enum value will increase over time as new events are added to the libvirt API",
	}
)

type collectFunc func(ch chan<- prometheus.Metric, l *libvirt.Libvirt, domain domainMeta, promLabels []string, logger log.Logger) (err error)

type domainMeta struct {
	domainName      string
	instanceName    string
	instanceId      string
	flavorName      string
	os_type_arch    string
	os_type_machine string
	os_type         string

	userName string
	userId   string

	projectName string
	projectId   string

	libvirtDomain libvirt.Domain
	libvirtSchema libvirt_schema.Domain
}

// LibvirtExporter implements a Prometheus exporter for libvirt state.
type LibvirtExporter struct {
	uri    string
	driver libvirt.ConnectURI

	logger log.Logger
}

// NewLibvirtExporter creates a new Prometheus exporter for libvirt.
func NewLibvirtExporter(uri string, driver libvirt.ConnectURI, logger log.Logger) (*LibvirtExporter, error) {
	return &LibvirtExporter{
		uri:    uri,
		driver: driver,
		logger: logger,
	}, nil
}

// DomainFromLibvirt retrives all domains from the libvirt socket and enriches them with some meta information.
func DomainsFromLibvirt(l *libvirt.Libvirt, logger log.Logger) ([]domainMeta, error) {
	domains, _, err := l.ConnectListAllDomains(1, 0)
	if err != nil {
		_ = level.Error(logger).Log("err", "failed to load domains", "msg", err)
		return nil, err
	}

	lvDomains := make([]domainMeta, len(domains))
	for idx, domain := range domains {
		xmlDesc, err := l.DomainGetXMLDesc(domain, 0)
		if err != nil {
			_ = level.Error(logger).Log("err", "failed to DomainGetXMLDesc", "domain", "instance_name", "project_id", "project_name", domain.Name, "msg", err)
			continue
		}
		var libvirtSchema libvirt_schema.Domain
		if err = xml.Unmarshal([]byte(xmlDesc), &libvirtSchema); err != nil {
			_ = level.Error(logger).Log("err", "failed to unmarshal domain", "domain", "instance_name", "project_id", "project_name", domain.Name, "msg", err)
			continue
		}

		lvDomains[idx].libvirtDomain = domain
		lvDomains[idx].libvirtSchema = libvirtSchema

		lvDomains[idx].domainName = domain.Name
		lvDomains[idx].instanceName = libvirtSchema.Metadata.NovaInstance.Name
		lvDomains[idx].instanceId = libvirtSchema.UUID
		lvDomains[idx].flavorName = libvirtSchema.Metadata.NovaInstance.Flavor.FlavorName
		lvDomains[idx].os_type_arch = libvirtSchema.OSMetadata.Type.Arch
		lvDomains[idx].os_type_machine = libvirtSchema.OSMetadata.Type.Machine
		lvDomains[idx].os_type = libvirtSchema.OSMetadata.Type.Value

		lvDomains[idx].userName = libvirtSchema.Metadata.NovaInstance.Owner.User.UserName
		lvDomains[idx].userId = libvirtSchema.Metadata.NovaInstance.Owner.User.UserId

		lvDomains[idx].projectName = libvirtSchema.Metadata.NovaInstance.Owner.Project.ProjectName
		lvDomains[idx].projectId = libvirtSchema.Metadata.NovaInstance.Owner.Project.ProjectId
	}

	return lvDomains, nil
}

// Collect scrapes Prometheus metrics from libvirt.
func (e *LibvirtExporter) Collect(ch chan<- prometheus.Metric) {
	if err := CollectFromLibvirt(ch, e.uri, e.driver, e.logger); err != nil {
		_ = level.Error(e.logger).Log("err", "failed to collect metrics", "msg", err)
	}
}


// CollectFromLibvirt obtains Prometheus metrics from all domains in a libvirt setup.
func CollectFromLibvirt(ch chan<- prometheus.Metric, uri string, driver libvirt.ConnectURI, logger log.Logger) (err error) {
	dialer := dialers.NewLocal(dialers.WithSocket(uri), dialers.WithLocalTimeout((5 * time.Second)))
	l := libvirt.NewWithDialer(dialer)
	if err = l.ConnectToURI(driver); err != nil {
		_ = level.Error(logger).Log("err", "failed to connect", "msg", err)
		return err
	}

	defer func() {
		if err := l.Disconnect(); err != nil {
			_ = level.Error(logger).Log("err", "failed to disconnect", "msg", err)
		}
	}()

	ch <- prometheus.MustNewConstMetric(
		libvirtUpDesc,
		prometheus.GaugeValue,
		1.0)

	domains, err := DomainsFromLibvirt(l, logger)
	if err != nil {
		_ = level.Error(logger).Log("err", "failed to retrieve domains from Libvirt", "msg", err)
		return err
	}


	domainNumber := len(domains)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainNumbers,
		prometheus.GaugeValue,
		float64(domainNumber))

	// collect domain metrics from libvirt
	// see https://libvirt.org/html/libvirt-libvirt-domain.html
	for _, domain := range domains {
		if err = CollectDomain(ch, l, domain, logger); err != nil {
			_ = level.Error(logger).Log("err", "failed to collect domain", "domain", "instance_name", "project_id", "project_name", domain.domainName, "msg", err)
			return err
		}
	}

	// collect storage pool metrics
	// see https://libvirt.org/html/libvirt-libvirt-storage.html
	var pools []libvirt.StoragePool
	if pools, _, err = l.ConnectListAllStoragePools(1, 0); err != nil {
		_ = level.Error(logger).Log("err", "failed to collect storage pools", "msg", err)
		return err
	}
	for _, pool := range pools {
		if err = CollectStoragePoolInfo(ch, l, pool, logger); err != nil {
			_ = level.Error(logger).Log("err", "failed to collect storage pool info", "msg", err)
			return err
		}
	}

	return nil
}

var rmaxmem, rmemory uint64

// CollectDomain extracts Prometheus metrics from a libvirt domain.
func CollectDomain(ch chan<- prometheus.Metric, l *libvirt.Libvirt, domain domainMeta, logger log.Logger) (err error) {

	var rState uint8
	var rvirCpu uint16
	var rcputime uint64
	if rState, rmaxmem, rmemory, rvirCpu, rcputime, err = l.DomainGetInfo(domain.libvirtDomain); err != nil {
		_ = level.Error(logger).Log("err", "failed to get domainInfo", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
		return err
	}

	promLabels := []string{
		domain.domainName,
		domain.instanceName,
		domain.projectName,
                domain.projectId,
	}

	openstackInfoLabels := []string{
		domain.domainName,
		domain.instanceName,
		domain.instanceId,
		domain.flavorName,
		domain.userName,
		domain.userId,
		domain.projectName,
		domain.projectId,
	}

	infoLabels := []string{
		domain.domainName,
		domain.instanceName,
                domain.projectName,
                domain.projectId,
		domain.os_type,
		domain.os_type_arch,
		domain.os_type_machine,
	}

	ch <- prometheus.MustNewConstMetric(libvirtDomainInfoDesc, prometheus.GaugeValue, 1.0, infoLabels...)
	ch <- prometheus.MustNewConstMetric(libvirtDomainOpenstackInfoDesc, prometheus.GaugeValue, 1.0, openstackInfoLabels...)

	ch <- prometheus.MustNewConstMetric(libvirtDomainState, prometheus.GaugeValue, float64(rState), append(promLabels, domainState[libvirt_schema.DomainState(rState)])...)

	ch <- prometheus.MustNewConstMetric(libvirtDomainInfoMaxMemDesc, prometheus.GaugeValue, float64(rmaxmem)*1024, promLabels...)
	ch <- prometheus.MustNewConstMetric(libvirtDomainInfoMemoryDesc, prometheus.GaugeValue, float64(rmemory)*1024, promLabels...)
	ch <- prometheus.MustNewConstMetric(libvirtDomainInfoNrVirtCpuDesc, prometheus.GaugeValue, float64(rvirCpu), promLabels...)
	ch <- prometheus.MustNewConstMetric(libvirtDomainInfoCpuTimeDesc, prometheus.CounterValue, float64(rcputime)/1e9, promLabels...)

	var isActive int32
	if isActive, err = l.DomainIsActive(domain.libvirtDomain); err != nil {
		_ = level.Error(logger).Log("err", "failed to get active status of domain", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
		return err
	}
	if isActive != 1 {
		_ = level.Debug(logger).Log("debug", "domain is not active, skipping", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name)
		return nil
	}

	for _, collectFunc := range []collectFunc{CollectDomainBlockDeviceInfo, CollectDomainNetworkInfo, CollectDomainMemoryStatInfo, CollectDomainVCPUInfo} {
		if err = collectFunc(ch, l, domain, promLabels, logger); err != nil {
			_ = level.Warn(logger).Log("warn", "failed to collect some domain info", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
		}
	}

	return nil
}

type diskCache struct {
    ReadBytes float64
    ReadRequests float64
    WriteBytes float64
    WriteRequests float64
    Timestamp time.Time
}
var diskTimeCache = make(map[string]diskCache)

func CollectDomainBlockDeviceInfo(ch chan<- prometheus.Metric, l *libvirt.Libvirt, domain domainMeta, promLabels []string, logger log.Logger) (err error) {

	// Report block device statistics.

	for _, disk := range domain.libvirtSchema.Devices.Disks {
		if disk.Device == "cdrom" || disk.Device == "fd" {
			continue
		}

		var rRdReq, rRdBytes, rWrReq, rWrBytes int64
		if rRdReq, rRdBytes, rWrReq, rWrBytes, _, err = l.DomainBlockStats(domain.libvirtDomain, disk.Target.Device); err != nil {
			_ = level.Warn(logger).Log("warn", "failed to get DomainBlockStats", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
			return err
		}

		promDiskLabels := append(promLabels, disk.Target.Device)
		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockStatsRdBytesDesc,
			prometheus.CounterValue,
			float64(rRdBytes),
			promDiskLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockStatsRdReqDesc,
			prometheus.CounterValue,
			float64(rRdReq),
			promDiskLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockStatsWrBytesDesc,
			prometheus.CounterValue,
			float64(rWrBytes),
			promDiskLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockStatsWrReqDesc,
			prometheus.CounterValue,
			float64(rWrReq),
			promDiskLabels...)

	        _, capacityBytes, _, err := l.DomainGetBlockInfo(domain.libvirtDomain, disk.Target.Device, 0)
		if err != nil {
                        _ = level.Warn(logger).Log("warn", "failed to get BlockInfo", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
                        return err
                }

                blockIOTune, _, err := l.DomainGetBlockIOTune(domain.libvirtDomain, libvirt.OptString{disk.Target.Device}, 30, 0)
                if err != nil {
                        _ = level.Warn(logger).Log("warn", "failed to get DomainBlockIOTune", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
                        return err
                }


                blockStats, _, err := l.DomainBlockStatsFlags(domain.libvirtDomain, disk.Target.Device, 10, 0)
		if err != nil {
			_ = level.Warn(logger).Log("warn", "failed to get DomainBlockStats", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
                        return err
	        }

		var totalBytesSec, readBytesSec, writeBytesSec, totalIopsSec, writeIopsSec, readIopsSec float64
                for _, param := range blockIOTune {
                       if param.Field == "total_bytes_sec" {
			   totalBytesSec = float64(param.Value.I.(uint64)) // Convert to float64 for Prometheus
                       }
		       if param.Field == "read_bytes_sec" {
			       readBytesSec = float64(param.Value.I.(uint64)) // Convert to float64 for Prometheus
                       }
		       if param.Field == "write_bytes_sec" {
			       writeBytesSec = float64(param.Value.I.(uint64)) // Convert to float64 for Prometheus
                       }
		       if param.Field == "total_iops_sec" {
			       totalIopsSec = float64(param.Value.I.(uint64)) // Convert to float64 for Prometheus
                       }
		       if param.Field == "read_iops_sec" {
			       readIopsSec = float64(param.Value.I.(uint64)) // Convert to float64 for Prometheus
                       }
		       if param.Field == "write_iops_sec" {
			       writeIopsSec = float64(param.Value.I.(uint64)) // Convert to float64 for Prometheus
                       }
                }

		var readTotalTime, writeTotalTime float64
		for _, param := range blockStats {
                       if param.Field == "rd_total_times" {
                           readTotalTime = float64(param.Value.I.(int64)) // Convert to float64 for Prometheus
                       }
                       if param.Field == "wr_total_times" {
                               writeTotalTime = float64(param.Value.I.(int64)) // Convert to float64 for Prometheus
                       }
		}

		ch <- prometheus.MustNewConstMetric(
                        libvirtDomainBlockCapacityBytesDesc,
                        prometheus.GaugeValue,
                        float64(capacityBytes),
                        promDiskLabels...)


                // Throughput limits (bytes/sec)
                ch <- prometheus.MustNewConstMetric(
                        libvirtDomainBlockTotalBytesSecDesc,
			prometheus.GaugeValue,
                        totalBytesSec,
                        promDiskLabels...)

                ch <- prometheus.MustNewConstMetric(
                        libvirtDomainBlockWriteBytesSecDesc,
                        prometheus.GaugeValue,
                        float64(writeBytesSec),
                        promDiskLabels...)

                ch <- prometheus.MustNewConstMetric(
                        libvirtDomainBlockReadBytesSecDesc,
                        prometheus.GaugeValue,
                        float64(readBytesSec),
                        promDiskLabels...)

                // IOPS limits
                ch <- prometheus.MustNewConstMetric(
                        libvirtDomainBlockTotalIopsSecDesc,
                        prometheus.GaugeValue,
                        float64(totalIopsSec),
                        promDiskLabels...)

                ch <- prometheus.MustNewConstMetric(
                        libvirtDomainBlockWriteIopsSecDesc,
                        prometheus.GaugeValue,
                        float64(writeIopsSec),
                        promDiskLabels...)

                ch <- prometheus.MustNewConstMetric(
                        libvirtDomainBlockReadIopsSecDesc,
                        prometheus.GaugeValue,
                        float64(readIopsSec),
                        promDiskLabels...)
	        // Total Read/Write time
		ch <- prometheus.MustNewConstMetric(
                        libvirtDomainBlockRdTotalTimeSecondsDesc,
                        prometheus.CounterValue,
                        float64(readTotalTime),
                        promDiskLabels...)
		ch <- prometheus.MustNewConstMetric(
                        libvirtDomainBlockWrTotalTimeSecondsDesc,
                        prometheus.CounterValue,
                        float64(writeTotalTime),
                        promDiskLabels...)
		//Disk Usage Percent
		currentTime := time.Now()

		cached, exists := diskTimeCache[domain.domainName]
		if !exists {
			diskTimeCache[domain.domainName] = diskCache{
				ReadBytes:  float64(rRdBytes),
				ReadRequests: float64(rRdReq),
				WriteBytes: float64(rWrBytes),
				WriteRequests: float64(rWrReq),
				Timestamp:  currentTime,
			}
			return err
		}
		timeDelta := currentTime.Sub(cached.Timestamp).Seconds()
		deltaReadBytes := float64(rRdBytes) - cached.ReadBytes
		deltaReadRequests := float64(rRdReq) - cached.ReadRequests
		deltaWriteBytes := float64(rWrBytes) - cached.WriteBytes
		deltaWriteRequests := float64(rWrReq) - cached.WriteRequests
		deltaTotalRequests := deltaReadRequests + deltaWriteRequests 

		diskTimeCache[domain.domainName] = diskCache{
			ReadBytes:  float64(rRdBytes),
			ReadRequests: float64(rRdReq),
			WriteBytes: float64(rWrBytes),
			WriteRequests: float64(rWrReq),
			Timestamp:  currentTime,
		}

		if readIopsSec != 0 {
                        diskReadRequestsSec := float64(deltaReadRequests) / timeDelta
                        diskReadRequestsPercent := ( float64(diskReadRequestsSec) / readIopsSec ) * float64(100)

                        ch <- prometheus.MustNewConstMetric(
                                libvirtDomainBlockReadRequestsPercentDesc,
                                prometheus.GaugeValue,
                                float64(diskReadRequestsPercent),
                                promDiskLabels...)
                }
		if writeIopsSec != 0 {
                        diskWriteRequestsSec := float64(deltaWriteRequests) / timeDelta
                        diskWriteRequestsPercent := ( float64(diskWriteRequestsSec) / writeIopsSec ) * float64(100)

                        ch <- prometheus.MustNewConstMetric(
                                libvirtDomainBlockWriteRequestsPercentDesc,
                                prometheus.GaugeValue,
                                float64(diskWriteRequestsPercent),
                                promDiskLabels...)
                }
		if totalIopsSec != 0 {
                        diskTotalRequestsSec := float64(deltaTotalRequests) / timeDelta
                        diskTotalRequestsPercent := ( float64(diskTotalRequestsSec) / totalIopsSec ) * float64(100)

                        ch <- prometheus.MustNewConstMetric(
                                libvirtDomainBlockTotalRequestsPercentDesc,
                                prometheus.GaugeValue,
                                float64(diskTotalRequestsPercent),
                                promDiskLabels...)
                }
		if readBytesSec != 0 {
			diskReadBytesSec := float64(deltaReadBytes) / timeDelta
			diskReadBytesPercent := ( float64(diskReadBytesSec) / readBytesSec ) * float64(100)

                        ch <- prometheus.MustNewConstMetric(
                                libvirtDomainBlockReadBytesPercentDesc,
                                prometheus.GaugeValue,
                                float64(diskReadBytesPercent),
                                promDiskLabels...)
                }

		if writeBytesSec != 0 {
			diskWriteBytesSec := float64(deltaWriteBytes) / timeDelta
			diskWriteBytesPercent := ( float64(diskWriteBytesSec) / writeBytesSec ) * float64(100)

                        ch <- prometheus.MustNewConstMetric(
                                libvirtDomainBlockWriteBytesPercentDesc,
                                prometheus.GaugeValue,
                                float64(diskWriteBytesPercent),
                                promDiskLabels...)
                }
		if totalBytesSec != 0 {
			deltaTotalBytes := deltaReadBytes + deltaWriteBytes
			diskTotalBytesPercent := ( deltaTotalBytes / totalBytesSec )  * float64(100)
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockTotalBytesPercentDesc,
				prometheus.GaugeValue,
				float64(diskTotalBytesPercent),
                                promDiskLabels...)
		}


		promDiskInfoLabels := append(promLabels, disk.Type, disk.Target.Bus, disk.Driver.Name, disk.Driver.Type, disk.Driver.Cache, disk.Driver.Discard, disk.Source.File, disk.Source.Protocol, disk.Target.Device, disk.Serial)
		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockStatsInfo,
			prometheus.GaugeValue,
			float64(1),
			promDiskInfoLabels...)
	}
	return
}

func CollectDomainNetworkInfo(ch chan<- prometheus.Metric, l *libvirt.Libvirt, domain domainMeta, promLabels []string, logger log.Logger) (err error) {

	// Report network interface statistics.
	for _, iface := range domain.libvirtSchema.Devices.Interfaces {
		if iface.Target.Device == "" {
			continue
		}
		var rRxBytes, rRxPackets, rRxErrs, rRxDrop, rTxBytes, rTxPackets, rTxErrs, rTxDrop int64
		if rRxBytes, rRxPackets, rRxErrs, rRxDrop, rTxBytes, rTxPackets, rTxErrs, rTxDrop, err = l.DomainInterfaceStats(domain.libvirtDomain, iface.Target.Device); err != nil {
			_ = level.Warn(logger).Log("warn", "failed to get DomainInterfaceStats", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
			return err
		}
		newAliasName := strings.Replace(iface.Alias.Name, "net", "eth", 1)

		promInterfaceLabels := append(promLabels, iface.Target.Device, newAliasName)
		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceRxBytesDesc,
			prometheus.CounterValue,
			float64(rRxBytes),
			promInterfaceLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceRxPacketsDesc,
			prometheus.CounterValue,
			float64(rRxPackets),
			promInterfaceLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceRxErrsDesc,
			prometheus.CounterValue,
			float64(rRxErrs),
			promInterfaceLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceRxDropDesc,
			prometheus.CounterValue,
			float64(rRxDrop),
			promInterfaceLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceTxBytesDesc,
			prometheus.CounterValue,
			float64(rTxBytes),
			promInterfaceLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceTxPacketsDesc,
			prometheus.CounterValue,
			float64(rTxPackets),
			promInterfaceLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceTxErrsDesc,
			prometheus.CounterValue,
			float64(rTxErrs),
			promInterfaceLabels...)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceTxDropDesc,
			prometheus.CounterValue,
			float64(rTxDrop),
			promInterfaceLabels...)

		promInterfaceInfoLabels := append(promLabels, iface.Type, iface.Source.Bridge, iface.Target.Device, iface.MAC.Address, iface.Model.Type, iface.MTU.Size, newAliasName)
		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceInfo,
			prometheus.GaugeValue,
			float64(1),
			promInterfaceInfoLabels...)
	}
	return
}

func CollectDomainMemoryStatInfo(ch chan<- prometheus.Metric, l *libvirt.Libvirt, domain domainMeta, promLabels []string, logger log.Logger) (err error) {
	//collect stat info
	var rStats []libvirt.DomainMemoryStat
	if rStats, err = l.DomainMemoryStats(domain.libvirtDomain, uint32(libvirt.DomainMemoryStatNr), 0); err != nil {
		_ = level.Warn(logger).Log("warn", "failed to get DomainMemoryStats", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
		return err
	}

	var freeMemoryBytes, diskCached uint64
	for _, stat := range rStats {
		switch stat.Tag {
		case int32(libvirt.DomainMemoryStatSwapIn):
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainMemoryStatsSwapInBytesDesc,
				prometheus.GaugeValue,
				float64(stat.Val)*1024,
				promLabels...)
		case int32(libvirt.DomainMemoryStatSwapOut):
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainMemoryStatsSwapOutBytesDesc,
				prometheus.GaugeValue,
				float64(stat.Val)*1024,
				promLabels...)
		case int32(libvirt.DomainMemoryStatUnused):
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainMemoryStatsUnusedBytesDesc,
				prometheus.GaugeValue,
				float64(stat.Val*1024),
				promLabels...)
			freeMemoryBytes = stat.Val*1024
		case int32(libvirt.DomainMemoryStatAvailable):
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainMemoryStatsAvailableInBytesDesc,
				prometheus.GaugeValue,
				float64(stat.Val*1024),
				promLabels...)
		case int32(libvirt.DomainMemoryStatUsable):
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainMemoryStatsUsableBytesDesc,
				prometheus.GaugeValue,
				float64(stat.Val*1024),
				promLabels...)
		case int32(libvirt.DomainMemoryStatRss):
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainMemoryStatsRssBytesDesc,
				prometheus.GaugeValue,
				float64(stat.Val*1024),
				promLabels...)
		case int32(libvirt.DomainMemoryStatDiskCaches):
			diskCached = stat.Val*1024
                        ch <- prometheus.MustNewConstMetric(
                                libvirtDomainMemoryStatDiskCachesBytesDesc,
                                prometheus.GaugeValue,
                                float64(stat.Val*1024),
                                promLabels...)
                }
	}
			var freeMemoryPercent, usedMemoryPercent float64
			maxMemoryBytes := float64(rmaxmem)*1024
			freeMemoryPercent = (float64(freeMemoryBytes) / maxMemoryBytes) *float64(100)
			usedMemoryPercent = 100 - freeMemoryPercent
			usednocache :=     maxMemoryBytes - float64(freeMemoryBytes) - float64(diskCached)
			usednocachePercent := (float64(usednocache) / maxMemoryBytes) * float64(100)
                        ch <- prometheus.MustNewConstMetric(
                              libvirtDomainMemoryStatUsedPercentDesc,
                              prometheus.GaugeValue,
                              float64(usedMemoryPercent),
                              promLabels...)
			ch <- prometheus.MustNewConstMetric(
                              libvirtDomainMemoryStatFreePercentDesc,
                              prometheus.GaugeValue,
                              float64(freeMemoryPercent),
                              promLabels...)
			ch <- prometheus.MustNewConstMetric(
                              libvirtDomainMemoryStatUsednocachePercentDesc,
                              prometheus.GaugeValue,
                              float64(usednocachePercent),
                              promLabels...)
	return
}

// Cache to store the previous vCPU times and their timestamps
var vcpuTimeCache = make(map[string]map[int]vcpuCache)

type vcpuCache struct {
    lastTime uint64
    lastDelay uint64
    lastTimestamp time.Time
}

func CollectDomainVCPUInfo(ch chan<- prometheus.Metric, l *libvirt.Libvirt, domain domainMeta, promLabels []string, logger log.Logger) (err error) {
	//collect domain vCPU stats
	var stats []libvirt.DomainStatsRecord
	// ConnectGetAllDomainStats expects a list of domains
	var d []libvirt.Domain
	d = append(d, domain.libvirtDomain)

	if stats, err = l.ConnectGetAllDomainStats(d, uint32(libvirt.DomainStatsVCPU), 0); err != nil {
		_ = level.Warn(logger).Log("warn", "failed to get vcpu stats", "domain", "instance_name", "project_id", "project_name", domain.libvirtDomain.Name, "msg", err)
		return err
	}

	// Get domain name
	domainName := domain.libvirtDomain.Name
	//Initialize caching for Prometheus metrics
	if _, ok := vcpuTimeCache[domainName]; !ok {
		vcpuTimeCache[domainName] = make(map[int]vcpuCache)
	}


	current := regexp.MustCompile("vcpu.current")
	maximum := regexp.MustCompile("vcpu.maximum")
	vcpu_metrics := regexp.MustCompile(`vcpu\.\d+\.\w+`)
	for _, stat := range stats {
		for _, param := range stat.Params {
			switch true {
			case current.MatchString(param.Field):
				metric_value := param.Value.I.(uint32)
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainVCPUStatsCurrent,
					prometheus.GaugeValue,
					float64(metric_value),
					promLabels...)
			case maximum.MatchString(param.Field):
				metric_value := param.Value.I.(uint32)
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainVCPUStatsMaximum,
					prometheus.GaugeValue,
					float64(metric_value),
					promLabels...)
			case vcpu_metrics.MatchString(param.Field):
				r := regexp.MustCompile(`vcpu\.(\d+)\.(\w+)`)
				match := r.FindStringSubmatch(param.Field)
				promVCPULabels := append(promLabels, match[1])

				vcpuIndex, _ := strconv.Atoi(match[1])
				// Get the current time
                                currentTime := time.Now()

                                // Retrieve the cached data for this vCPU, if it exists
                                cached, exists := vcpuTimeCache[domainName][vcpuIndex]

				switch match[2] {
				case "state":
					metric_value := param.Value.I.(int32)
					ch <- prometheus.MustNewConstMetric(
						libvirtDomainVCPUStatsState,
						prometheus.GaugeValue,
						float64(metric_value),
						promVCPULabels...)
				case "time":
					metric_value := param.Value.I.(uint64)
					ch <- prometheus.MustNewConstMetric(
						libvirtDomainVCPUStatsTime,
						prometheus.CounterValue,
						float64(metric_value)/1e9,
						promVCPULabels...)

                                        if exists {
                                                // Calculate the time difference between the current and previous metric collection
                                                timeDelta := currentTime.Sub(cached.lastTimestamp).Seconds()

                                                if timeDelta > 0 {
                                                                 // Calculate the vCPU time delta
                                                                 cputimeDelta := float64(metric_value - cached.lastTime) / 1e9 // Convert nanoseconds to seconds

                                                                 // Calculate the CPU usage percentage over the measured interval
                                                                 cpuPercent := (cputimeDelta / timeDelta) * 100

                                                                 // Emit the metric for vCPU system percent
                                                                 promVCPULabels := append(promLabels, match[1])
                                                                 ch <- prometheus.MustNewConstMetric(
                                                                         libvirtDomainVCPUStatsSysPercent,
                                                                         prometheus.GaugeValue,
                                                                         cpuPercent,
                                                                         promVCPULabels...)
                                                }
                                        }

                                        // Store the current time and vCPU time value in the cache for future comparisons
                                        vcpuTimeCache[domainName][vcpuIndex] = vcpuCache{
                                                     lastTime:      metric_value,
						     lastDelay:     cached.lastDelay,
                                                     lastTimestamp: currentTime,
                                        }
				case "wait":
					metric_value := param.Value.I.(uint64)
					ch <- prometheus.MustNewConstMetric(
						libvirtDomainVCPUStatsWait,
						prometheus.CounterValue,
						float64(metric_value)/1e9,
						promVCPULabels...)
					// Calculate the delay (steal) metric similarly
                                        if exists {
                                                   timeDelta := currentTime.Sub(cached.lastTimestamp).Seconds()

                                                   if timeDelta > 0 {
                                                                    delayDelta := float64(metric_value - cached.lastDelay) / float64(1e9)

                                                                    // Calculate steal percent
                                                                    stealPercent := (delayDelta / timeDelta) * 100

                                                                    // Emit the metric for vCPU steal percent
                                                                    ch <- prometheus.MustNewConstMetric(
                                                                            libvirtDomainVCPUStatsStealPercent,
                                                                            prometheus.GaugeValue,
                                                                            stealPercent,
                                                                            promVCPULabels...)
				                   }
                                                   // Store the current time and vCPU time value in the cache for future comparisons
                                                   vcpuTimeCache[domainName][vcpuIndex] = vcpuCache{
                                                               lastTime:      cached.lastTime,
                                                               lastDelay:     metric_value,
                                                               lastTimestamp: currentTime,
					           }
				        }
				case "delay":
					metric_value := param.Value.I.(uint64)
					ch <- prometheus.MustNewConstMetric(
						libvirtDomainVCPUStatsDelay,
						prometheus.CounterValue,
						float64(metric_value)/1e9,
						promVCPULabels...)
				}
			}
		}
	}
	return
}

func CollectStoragePoolInfo(ch chan<- prometheus.Metric, l *libvirt.Libvirt, pool libvirt.StoragePool, logger log.Logger) (err error) {
	// Report storage pool metrics
	var rState uint8
	var rCapacity, rAllocation, rAvailable uint64

	promLabels := []string{
		pool.Name,
	}
	if rState, rCapacity, rAllocation, rAvailable, err = l.StoragePoolGetInfo(pool); err != nil {
		_ = level.Warn(logger).Log("warn", "failed to get StoragePoolInfo for pool", "pool", pool.Name, "msg", err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		libvirtStoragePoolState,
		prometheus.GaugeValue,
		float64(rState),
		promLabels...)
	ch <- prometheus.MustNewConstMetric(
		libvirtStoragePoolCapacity,
		prometheus.GaugeValue,
		float64(rCapacity),
		promLabels...)
	ch <- prometheus.MustNewConstMetric(
		libvirtStoragePoolAllocation,
		prometheus.GaugeValue,
		float64(rAllocation),
		promLabels...)
	ch <- prometheus.MustNewConstMetric(
		libvirtStoragePoolAvailable,
		prometheus.GaugeValue,
		float64(rAvailable),
		promLabels...)
	return
}

// Describe returns metadata for all Prometheus metrics that may be exported.
func (e *LibvirtExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- libvirtUpDesc
	ch <- libvirtDomainNumbers

	ch <- libvirtDomainInfoDesc
	ch <- libvirtDomainOpenstackInfoDesc

	//domain info
	ch <- libvirtDomainState
	ch <- libvirtDomainInfoMaxMemDesc
	ch <- libvirtDomainInfoMemoryDesc
	ch <- libvirtDomainInfoNrVirtCpuDesc
	ch <- libvirtDomainInfoCpuTimeDesc

	//domain block
	ch <- libvirtDomainBlockStatsInfo
	ch <- libvirtDomainBlockStatsRdBytesDesc
	ch <- libvirtDomainBlockStatsRdReqDesc
	ch <- libvirtDomainBlockStatsWrBytesDesc
	ch <- libvirtDomainBlockStatsWrReqDesc
	ch <- libvirtDomainBlockCapacityBytesDesc
	ch <- libvirtDomainBlockRdTotalTimeSecondsDesc
        ch <- libvirtDomainBlockWrTotalTimeSecondsDesc
	ch <- libvirtDomainBlockTotalBytesSecDesc
	ch <- libvirtDomainBlockReadBytesSecDesc
	ch <- libvirtDomainBlockWriteBytesSecDesc
	ch <- libvirtDomainBlockTotalIopsSecDesc
	ch <- libvirtDomainBlockReadIopsSecDesc
	ch <- libvirtDomainBlockWriteIopsSecDesc
	ch <- libvirtDomainBlockReadBytesPercentDesc
        ch <- libvirtDomainBlockWriteBytesPercentDesc
        ch <- libvirtDomainBlockTotalBytesPercentDesc
	ch <- libvirtDomainBlockReadRequestsPercentDesc
	ch <- libvirtDomainBlockWriteRequestsPercentDesc
	ch <- libvirtDomainBlockTotalRequestsPercentDesc

	//domain interface
	ch <- libvirtDomainInterfaceInfo
	ch <- libvirtDomainInterfaceRxBytesDesc
	ch <- libvirtDomainInterfaceRxPacketsDesc
	ch <- libvirtDomainInterfaceRxErrsDesc
	ch <- libvirtDomainInterfaceRxDropDesc
	ch <- libvirtDomainInterfaceTxBytesDesc
	ch <- libvirtDomainInterfaceTxPacketsDesc
	ch <- libvirtDomainInterfaceTxErrsDesc
	ch <- libvirtDomainInterfaceTxDropDesc

	//domain mem stat
	ch <- libvirtDomainMemoryStatsSwapInBytesDesc
	ch <- libvirtDomainMemoryStatsSwapOutBytesDesc
	ch <- libvirtDomainMemoryStatsUnusedBytesDesc
	ch <- libvirtDomainMemoryStatsAvailableInBytesDesc
	ch <- libvirtDomainMemoryStatsUsableBytesDesc
	ch <- libvirtDomainMemoryStatsRssBytesDesc
	ch <- libvirtDomainMemoryStatUsedPercentDesc
	ch <- libvirtDomainMemoryStatDiskCachesBytesDesc
	ch <- libvirtDomainMemoryStatUsedPercentDesc
	ch <- libvirtDomainMemoryStatFreePercentDesc
	ch <- libvirtDomainMemoryStatUsednocachePercentDesc

	//domain vcpu stats
	ch <- libvirtDomainVCPUStatsCurrent
	ch <- libvirtDomainVCPUStatsMaximum
	ch <- libvirtDomainVCPUStatsState
	ch <- libvirtDomainVCPUStatsTime
	ch <- libvirtDomainVCPUStatsWait
	ch <- libvirtDomainVCPUStatsDelay
	ch <- libvirtDomainVCPUStatsSysPercent
	ch <- libvirtDomainVCPUStatsStealPercent

	//storage pool metrics
	ch <- libvirtStoragePoolState
	ch <- libvirtStoragePoolCapacity
	ch <- libvirtStoragePoolAllocation
	ch <- libvirtStoragePoolAvailable
}
