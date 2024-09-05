package gonetplan

import "github.com/vishvananda/netlink"

func GetLinks() ([]netlink.Link, error) {
	return netlink.LinkList()
}

func GetEthernetLinks() ([]netlink.Link, error) {
	links, err := GetLinks()
	if err != nil {
		return nil, err
	}

	ethLinks := []netlink.Link{}
	for _, link := range links {
		if link.Type() == "device" {
			ethLinks = append(ethLinks, link)
		}
	}

	return ethLinks, nil
}

func GetVlanLinks() ([]netlink.Link, error) {
	links, err := GetLinks()
	if err != nil {
		return nil, err
	}

	vlanLinks := []netlink.Link{}
	for _, link := range links {
		if link.Type() == "vlan" {
			vlanLinks = append(vlanLinks, link)
		}
	}

	return vlanLinks, nil
}

func GetBridgeLinks() ([]netlink.Link, error) {
	links, err := GetLinks()
	if err != nil {
		return nil, err
	}

	bridgeLinks := []netlink.Link{}
	for _, link := range links {
		if link.Type() == "bridge" {
			bridgeLinks = append(bridgeLinks, link)
		}
	}

	return bridgeLinks, nil
}

func GetBondLinks() ([]netlink.Link, error) {
	links, err := GetLinks()
	if err != nil {
		return nil, err
	}

	bondLinks := []netlink.Link{}
	for _, link := range links {
		if link.Type() == "bond" {
			bondLinks = append(bondLinks, link)
		}
	}

	return bondLinks, nil
}

func GetBondBridgeVlanNames() []string {
	links, _ := GetLinks()
	names := []string{}
	for _, link := range links {
		if link.Type() == "bond" || link.Type() == "bridge" || link.Type() == "vlan" {
			names = append(names, link.Attrs().Name)
		}
	}
	return names
}
