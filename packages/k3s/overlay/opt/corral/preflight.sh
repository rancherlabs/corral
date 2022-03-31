#!/bin/bash

echo "$CORRAL_corral_user_public_key" >> /$(whoami)/.ssh/authorized_keys
