
# Prometheus-libvirt-exporter
[![Build and Test](https://github.com/inovex/prometheus-libvirt-exporter/actions/workflows/build_and_test.yml/badge.svg)](https://github.com/inovex/prometheus-libvirt-exporter/actions/workflows/build_and_test.yml)
[![Lint Go Code](https://github.com/inovex/prometheus-libvirt-exporter/actions/workflows/lint.yml/badge.svg)](https://github.com/inovex/prometheus-libvirt-exporter/actions/workflows/lint.yml)

A prometheus-[libvirt](https://libvirt.org/)-exporter for host and vm metrics exposed for prometheus, written in Go with pluggable metric collectors.
By default, this exporter listens on TCP port 9177, path '/metrics', to expose metrics.

This exporter is built upon the [go-libvirt](https://github.com/digitalocean/go-libvirt) package developed by DigitalOcean. It offers a pure Go interface for interacting with Libvirt, leveraging the RPC interface provided by Libvirt. For detailed information about the Go bindings used, you can refer to the [Libvirt API reference](https://libvirt.org/html/index.html).



# Building and running

This release provides a set of assets for the prometheus-libvirt-exporter. It includes installation packages for various platforms (apk, deb, rpm) and the the binaries. Additionally, source code archives in both zip and tar.gz formats are available for download.

## Requirements
1. Gorelease: `go install github.com/goreleaser/goreleaser@latest`

2. Taskfile: `go install github.com/go-task/task/v3/cmd/task@latest`

## Local Building
1. Run `task build`

2. Afterwards all packages, binaries and archives are available in the `dist/` folder

## To see all available configuration flags:

`./prometheus-libvirt-exporter -h`


## metrics
Name | Label |Description
---------|---------|-------------
up||scraping libvirt's metrics state
libvirt_domains||number of domains
libvirt_domain_openstack_info | "domain", "instance_name", "instance_id", "flavor_name", "user_name", "user_id", "project_name", "project_id" | Aggregated OpenStack metadata as labels
libvirt_domain_info | "project_name", "project_id", "domain", "instance_name", "os_type", "os_type_machine", "os_type_arch" | e.g. os (operating system booting) settings as labels
libvirt_domain_info_state | "project_name", "project_id", "domain", "instance_name", "state_desc" | Code of the domain state, include state description
libvirt_domain_info_maximum_memory_bytes | "project_name", "project_id", "domain", "instance_name" | Maximum allowed memory of the domain
libvirt_domain_info_memory_usage_bytes | "project_name", "project_id", "domain", "instance_name" | Memory usage of the domain
libvirt_domain_info_virtual_cpus | "project_name", "project_id", "domain", "instance_name" | Number of virtual CPUs for the domain
libvirt_domain_info_cpu_time_seconds_total | "project_name", "project_id", "domain", "instance_name" | Amount of CPU time used by the domain
libvirt_domain_memory_stats_swap_in_bytes | "project_name", "project_id", "domain", "instance_name" | Memory swapped in for this domain (the total amount of data read from swap space)
libvirt_domain_memory_stats_swap_out_bytes | "project_name", "project_id", "domain", "instance_name" | Memory swapped out for this domain (the total amount of memory written out to swap space)
libvirt_domain_memory_stats_unused_bytes | "project_name", "project_id", "domain", "instance_name" | Memory unused by the domain
libvirt_domain_memory_stats_available_bytes | "project_name", "project_id", "domain", "instance_name" | Memory available to the domain
libvirt_domain_memory_stats_usable_bytes | "project_name", "project_id", "domain", "instance_name" | Memory usable by the domain (corresponds to 'Available' in /proc/meminfo)
libvirt_domain_memory_stats_rss_bytes | "project_name", "project_id", "domain", "instance_name" | Resident Set Size of the process running the domain
libvirt_domain_memory_stats_disk_cache_bytes | "project_name", "project_id", "domain", "instance_name" | The amount of memory that can be quickly reclaimed without additional I/O (in bytes).
libvirt_domain_memory_stats_used_percent | "project_name", "project_id", "domain", "instance_name" | The amount of memory in percent that is used by the domain.
libvirt_domain_memory_stats_free_percent | "project_name", "project_id", "domain", "instance_name" | The percentage of memory currently available for use by the instance
libvirt_domain_memory_stats_usednocache_percent | "project_name", "project_id", "domain", "instance_name" | The percentage of memory currently used without page cache/buffer cache by the instance
libvirt_domain_block_stats_info | "project_name", "project_id", "domain", "instance_name", "disk_type", "driver_cache", "driver_discard", "driver_name", "driver_type", "serial", "source_file", "target_bus", "target_device" | Metadata information on block devices
libvirt_domain_block_stats_read_bytes_total | "project_name", "project_id", "domain", "instance_name", "target_device", "host" | Number of bytes read from a block device, in bytes
libvirt_domain_block_stats_read_requests_total | "project_name", "project_id", "domain", "instance_name", "target_device", "host" | Number of read requests from a block device
libvirt_domain_block_stats_write_bytes_total | "project_name", "project_id", "domain", "instance_name", "target_device" | Number of bytes written from a block device, in bytes
libvirt_domain_block_stats_write_requests_total | "project_name", "project_id", "domain", "instance_name", "target_device" | Number of write requests from a block device
libvirt_domain_block_stats_limit_total_bytes | "project_name", "project_id", "domain", "instance_name", "target_device" | Total throughput limit in bytes per second 
libvirt_domain_block_stats_limit_total_requests | "project_name", "project_id", "domain", "instance_name", "target_device" | Total requests limit in bytes per second
libvirt_domain_block_stats_limit_read_bytes | "project_name", "project_id", "domain", "instance_name", "target_device" | Read throughput limit in bytes per second 
libvirt_domain_block_stats_limit_read_requests | "project_name", "project_id", "domain", "instance_name", "target_device" | Read requests limit in bytes per second
libvirt_domain_block_stats_limit_write_bytes | "project_name", "project_id", "domain", "instance_name", "target_device" | Write throughput limit in bytes per second
libvirt_domain_block_stats_limit_write_requests | "project_name", "project_id", "domain", "instance_name", "target_device" | Write requests limit in bytes per second
libvirt_domain_block_stats_capacity_bytes | "project_name", "project_id", "domain", "instance_name", "target_device" | Logical size in bytes of the block device
libvirt_domain_block_stats_read_bytes_usage_percent | "project_name", "project_id", "domain", "instance_name", "target_device" | Read bytes usage percent
libvirt_domain_block_stats_write_bytes_usage_percent | "project_name", "project_id", "domain", "instance_name", "target_device" | Write bytes usage percent
libvirt_domain_block_stats_total_bytes_usage_percent | "project_name", "project_id", "domain", "instance_name", "target_device" | Total bytes usage percent
libvirt_domain_block_stats_read_requests_usage_percent | "project_name", "project_id", "domain", "instance_name", "target_device" | Read requests usage percent
libvirt_domain_block_stats_write_requests_usage_percent | "project_name", "project_id", "domain", "instance_name", "target_device" | Write requests usage percent
libvirt_domain_block_stats_total_requests_usage_percent | "project_name", "project_id", "domain", "instance_name", "target_device" | Total requests usage percent
libvirt_domain_interface_stats_info | "project_name", "project_id", "domain", "instance_name", "alias_name", "interface_type", "mac_address", "model_type", "mtu_size", "source_bridge", "target_device" | Metadata on network interfaces
libvirt_domain_interface_stats_receive_bytes_total | "project_name", "project_id", "domain", "instance_name", "alias_name", "target_device" | Number of bytes received on a network interface, in bytes
libvirt_domain_interface_stats_receive_packets_total | "project_name", "project_id", "domain", "instance_name", "alias_name", "target_device" | Number of packets received on a network interface
libvirt_domain_interface_stats_receive_errors_total | "project_name", "project_id", "domain", "instance_name", "alias_name", "target_device" | Number of packet receive errors on a network interface
libvirt_domain_interface_stats_receive_drops_total | "project_name", "project_id", "domain", "instance_name", "alias_name", "target_device" | Number of packet receive drops on a network interface
libvirt_domain_interface_stats_transmit_bytes_total | "project_name", "project_id", "domain", "instance_name", "alias_name", "target_device" | Number of bytes transmitted on a network interface, in bytes
libvirt_domain_interface_stats_transmit_packets_total | "project_name", "project_id", "domain", "instance_name", "alias_name", "target_device" | Number of packets transmitted on a network interface
libvirt_domain_interface_stats_transmit_errors_total | "project_name", "project_id", "domain", "instance_name", "alias_name", "target_device" | Number of packet transmit errors on a network interface
libvirt_domain_interface_stats_transmit_drops_total | "project_name", "project_id", "domain", "instance_name", "alias_name", "target_device" | Number of packet transmit drops on a network interface
libvirt_domain_vcpu_current | "project_name", "project_id", "domain", "instance_name" | Number of current online vCPUs
libvirt_domain_vcpu_delay_seconds_total | "project_name", "project_id", "domain", "instance_name", "vcpu" | Time the vCPU spent waiting in the queue instead of running. Exposed to the VM as steal time
libvirt_domain_vcpu_maximum | "project_name", "project_id", "domain", "instance_name" | Number of maximum online vCPUs
libvirt_domain_vcpu_state | "project_name", "project_id", "domain", "instance_name", "vcpu" | State of the vCPU
libvirt_domain_vcpu_time_seconds_total | "project_name", "project_id", "domain", "instance_name", "vcpu" | Time spent by the virtual CPU
libvirt_domain_vcpu_wait_seconds_total | "project_name", "project_id", "domain", "instance_name", "vcpu" | Time the vCPU wants to run, but the host scheduler has something else running ahead of it
libvirt_domain_vcpu_sys_percent | "project_name", "project_id", "domain", "instance_name", "vcpu" | CPU usage percent by instance on all vCPUs 
libvirt_domain_vcpu_steal_percent | "project_name", "project_id", "domain", "instance_name", "vcpu" | The percentage of time the virtual machine process is waiting on the physical CPU for its CPU time
libvirt_domain_storage_pool_allocation_bytes | "storage_pool" | Current allocation bytes of the storage pool
libvirt_domain_storage_pool_available_bytes | "storage_pool" | Remaining free space of the storage pool in bytes
libvirt_domain_storage_pool_capacity_bytes | "storage_pool" | Size of the storage pool in logical bytes
libvirt_domain_storage_pool_state | "storage_pool" | State of the storage pool



## Example

```
instance_block_stats_capacity_bytes{domain="instance-00009c24",target_device="sda"} 5.36870912e+10
instance_block_stats_info{disk_type="network",domain="instance-00009c24",driver_cache="none",driver_discard="unmap",driver_name="qemu",driver_type="raw",serial="92418ed9-5ada-4aab-8310-44c0ea26d9b2",source_file="",source_protocol="rbd",target_bus="scsi",target_device="sda"} 1
instance_block_stats_limit_read_bytes{domain="instance-00009c24",target_device="sda"} 2.097152e+08
instance_block_stats_limit_read_requests{domain="instance-00009c24",target_device="sda"} 2000
instance_block_stats_limit_total_bytes{domain="instance-00009c24",target_device="sda"} 0
instance_block_stats_limit_total_requests{domain="instance-00009c24",target_device="sda"} 0
instance_block_stats_limit_write_bytes{domain="instance-00009c24",target_device="sda"} 1.048576e+08
instance_block_stats_limit_write_requests{domain="instance-00009c24",target_device="sda"} 1000
instance_block_stats_read_bytes_total{domain="instance-00009c24",target_device="sda"} 1.478780928e+09
instance_block_stats_read_requests_total{domain="instance-00009c24",target_device="sda"} 25352
instance_block_stats_read_time_seconds_total{domain="instance-00009c24",target_device="sda"} 2.3555064842e+10
instance_block_stats_write_bytes_total{domain="instance-00009c24",target_device="sda"} 8.548481792e+10
instance_block_stats_write_requests_total{domain="instance-00009c24",target_device="sda"} 1.8040082e+07
instance_block_stats_write_time_seconds_total{domain="instance-00009c24",target_device="sda"} 2.9585565532515e+13
instance_domain_info{domain="instance-00009c24",os_type="hvm",os_type_arch="x86_64",os_type_machine="pc-i440fx-4.2"} 1
instance_domain_openstack_info{domain="instance-00009c24",flavor_name="i-pro-small.2x4",instance_id="96ff470a-602d-47ae-b437-f013e36b49dd",instance_name="sb-prj-20240820-005-phuong1709-controlplane-v0-v4hqp",project_id="2c2f46c54aba476b8d95b54431b7c093",project_name="kaas",user_id="02fe00929267453497fab4ddbda53618",user_name="cloud_portal"} 1
instance_info_cpu_time_seconds_total{domain="instance-00009c24"} 248416.4
instance_info_maximum_memory_bytes{domain="instance-00009c24"} 4.294967296e+09
instance_info_memory_usage_bytes{domain="instance-00009c24"} 4.294967296e+09
instance_info_state{domain="instance-00009c24",state_desc="the domain is running"} 1
instance_info_virtual_cpus{domain="instance-00009c24"} 2
instance_interface_stats_info{domain="instance-00009c24",interface_type="bridge",mac_address="fa:16:3e:e7:14:b2",model_type="virtio",mtu_size="1450",source_bridge="brq16629583-1a",target_device="tap7cd5a4a6-19"} 1
instance_interface_stats_receive_bytes_total{domain="instance-00009c24",target_device="tap7cd5a4a6-19"} 1.1502659224e+10
instance_interface_stats_receive_drops_total{domain="instance-00009c24",target_device="tap7cd5a4a6-19"} 0
instance_interface_stats_receive_errors_total{domain="instance-00009c24",target_device="tap7cd5a4a6-19"} 0
instance_interface_stats_receive_packets_total{domain="instance-00009c24",target_device="tap7cd5a4a6-19"} 6.8510279e+07
instance_interface_stats_transmit_bytes_total{domain="instance-00009c24",target_device="tap7cd5a4a6-19"} 9.964156364e+09
instance_interface_stats_transmit_drops_total{domain="instance-00009c24",target_device="tap7cd5a4a6-19"} 0
instance_interface_stats_transmit_errors_total{domain="instance-00009c24",target_device="tap7cd5a4a6-19"} 0
instance_interface_stats_transmit_packets_total{domain="instance-00009c24",target_device="tap7cd5a4a6-19"} 6.689421e+07
instance_memory_stats_available_bytes{domain="instance-00009c24"} 4.100927488e+09
instance_memory_stats_disk_cache_bytes{domain="instance-00009c24"} 2.068062208e+09
instance_memory_stats_free_percent{domain="instance-00009c24"} 25.929737091064453
instance_memory_stats_rss_bytes{domain="instance-00009c24"} 2.749120512e+09
instance_memory_stats_swap_in_bytes{domain="instance-00009c24"} 0
instance_memory_stats_swap_out_bytes{domain="instance-00009c24"} 0
instance_memory_stats_unused_bytes{domain="instance-00009c24"} 1.113673728e+09
instance_memory_stats_usable_bytes{domain="instance-00009c24"} 2.99044864e+09
instance_memory_stats_used_percent{domain="instance-00009c24"} 74.07026290893555
instance_memory_stats_usednocache_percent{domain="instance-00009c24"} 25.919437408447266
instance_vcpu_current{domain="instance-00009c24"} 2
instance_vcpu_delay_seconds_total{domain="instance-00009c24",vcpu="0"} 291.394911295
instance_vcpu_delay_seconds_total{domain="instance-00009c24",vcpu="1"} 288.515926297
instance_vcpu_maximum{domain="instance-00009c24"} 2
instance_vcpu_state{domain="instance-00009c24",vcpu="0"} 1
instance_vcpu_state{domain="instance-00009c24",vcpu="1"} 1
instance_vcpu_steal_percent{domain="instance-00009c24",vcpu="0"} 0
instance_vcpu_steal_percent{domain="instance-00009c24",vcpu="1"} 0
instance_vcpu_sys_percent{domain="instance-00009c24",vcpu="0"} 12.285174195760163
instance_vcpu_sys_percent{domain="instance-00009c24",vcpu="1"} 13.51338936962444
instance_vcpu_time_seconds_total{domain="instance-00009c24",vcpu="0"} 115561.99
instance_vcpu_time_seconds_total{domain="instance-00009c24",vcpu="1"} 121254.38
instance_vcpu_wait_seconds_total{domain="instance-00009c24",vcpu="0"} 0
instance_vcpu_wait_seconds_total{domain="instance-00009c24",vcpu="1"} 0
```
