#!/bin/bash

echo "$CORRAL_corral_user_public_key" >> /$(whoami)/.ssh/authorized_keys
echo "[keyfile]
      unmanaged-devices=interface-name:cali*;interface-name:flannel*" > /etc/NetworkManager/conf.d/rke2-canal.conf

systemctl reload NetworkManager
