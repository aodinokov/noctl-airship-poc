#!/bin/bash
#functions
# Copyright 2018 AT&T Intellectual Property.  All other rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

function _debootstrap (){
  debootstrap \
    --arch=amd64 \
    --variant=minbase \
    --foreign \
    focal \
    "${HOME}"/LIVE_BOOT/chroot \
    http://archive.ubuntu.com/ubuntu/
  chroot "${HOME}"/LIVE_BOOT/chroot /debootstrap/debootstrap --second-stage
}

function _make_kernel(){
  mkdir -p "${HOME}"/LIVE_BOOT/{scratch,image/live}
  mksquashfs \
    "${HOME}"/LIVE_BOOT/chroot \
    "${HOME}"/LIVE_BOOT/image/live/filesystem.squashfs \
    -e boot

  cp "${HOME}"/LIVE_BOOT/chroot/boot/vmlinuz-* \
     "${HOME}"/LIVE_BOOT/image/vmlinuz &&
  cp "${HOME}"/LIVE_BOOT/chroot/boot/initrd.img-* \
     "${HOME}"/LIVE_BOOT/image/initrd
}

function _grub_install (){
  cp "${1:?}"/grub.conf "${HOME}"/LIVE_BOOT/scratch/grub.cfg

  touch "${HOME}/LIVE_BOOT/image/UBUNTU_FOCAL_CUSTOM"

  grub-mkstandalone \
    --format=x86_64-efi \
    --output="${HOME}/LIVE_BOOT/scratch/bootx64.efi" \
    --locales="" \
    --fonts="" \
    boot/grub/grub.cfg="${HOME}/LIVE_BOOT/scratch/grub.cfg"

  (
    cd "${HOME}/LIVE_BOOT/scratch" && \
    dd if=/dev/zero of=efiboot.img bs=1M count=10 && \
    mkfs.vfat efiboot.img && \
    LC_CTYPE=C mmd -i efiboot.img efi efi/boot && \
    LC_CTYPE=C mcopy -i efiboot.img ./bootx64.efi ::efi/boot/
  )

  grub-mkstandalone \
     --format=i386-pc \
     --output="${HOME}/LIVE_BOOT/scratch/core.img" \
     --install-modules="linux16 linux normal iso9660 biosdisk memdisk search tar ls" \
     --modules="linux16 linux normal iso9660 biosdisk search" \
     --locales="" \
     --fonts="" \
     boot/grub/grub.cfg="${HOME}/LIVE_BOOT/scratch/grub.cfg"

  cat \
      /usr/lib/grub/i386-pc/cdboot.img "${HOME}/LIVE_BOOT/scratch/core.img" \
      > "${HOME}/LIVE_BOOT/scratch/bios.img"
}

function _make_iso(){
  xorriso \
    -as mkisofs \
    -iso-level 3 \
    -full-iso9660-filenames \
    -volid "config-2" \
    -eltorito-boot boot/grub/bios.img \
    -no-emul-boot \
    -boot-load-size 4 \
    -boot-info-table \
    --eltorito-catalog boot/grub/boot.cat \
    --grub2-boot-info \
    --grub2-mbr /usr/lib/grub/i386-pc/boot_hybrid.img \
    -eltorito-alt-boot \
    -e EFI/efiboot.img \
    -no-emul-boot \
    -append_partition 2 0xef "${HOME}/LIVE_BOOT/scratch/efiboot.img" \
    -output "${1:?}/ubuntu-focal.iso" \
    -graft-points \
        "${HOME}/LIVE_BOOT/image" \
        /boot/grub/bios.img="${HOME}/LIVE_BOOT/scratch/bios.img" \
        /EFI/efiboot.img="${HOME}/LIVE_BOOT/scratch/efiboot.img"
}

function _make_metadata(){
  echo "bootImagePath: ${2:?}/ubuntu-focal.iso" > "${1:?}"
}

function _check_input_data_set_vars(){
  CHROOT="${HOME}/LIVE_BOOT/chroot"
  export CHROOT
  export CLOUD_DATA_LATEST="${HOME}/LIVE_BOOT/image/openstack/latest"
  echo "${BUILDER_CONFIG:?}"
  if [ ! -f "${BUILDER_CONFIG}" ]
  then
      echo "file ${BUILDER_CONFIG} not found"
      exit 1
  fi
  IFS=':' read -ra ADDR <<<"$(yq r "${BUILDER_CONFIG}" container.volume)"
  VOLUME="${ADDR[1]}"
  echo "${VOLUME:?}"
  if [[ "${VOLUME}" == 'none' ]]
  then
      echo "variable container.volume \
           is not present in $BUILDER_CONFIG"
      exit 1
  else
      if [[ ! -d "${VOLUME}" ]]
      then
          echo "${VOLUME} not exist"
          exit 1
      fi
  fi
  #USER_DATA="${VOLUME}/$(yq r "${BUILDER_CONFIG}" builder.userDataFileName)"
  USER_DATA=$(DIR=$(yq r "${BUILDER_CONFIG}" builder.userDataFileName) && [ "${DIR:0:1}" = "/" ] && echo "${DIR}" || echo "${VOLUME}/${DIR}")
  echo "${USER_DATA:?}"
  if [[ "${USER_DATA}" == 'none' ]]
  then
      echo "variable userDataFileName \
          is not present in ${BUILDER_CONFIG}"
      exit 1
  else
      if [[ ! -f ${USER_DATA} ]]
      then
          echo "${USER_DATA} not exist"
          exit 1
      fi
  fi
  #NET_CONFIG="${VOLUME}/$(yq r "${BUILDER_CONFIG}" \
  #    builder.networkConfigFileName)"
  NET_CONFIG=$(DIR=$(yq r "${BUILDER_CONFIG}" builder.networkConfigFileName) && [ "${DIR:0:1}" = "/" ] && echo "${DIR}" || echo "${VOLUME}/${DIR}")
  echo "${NET_CONFIG:?}"
  if [[ "${NET_CONFIG}" == 'none' ]]
  then
      echo "variable networkConfigFileName \
          is not present in ${BUILDER_CONFIG}"
      exit 1
      if [[ ! -f ${NET_CONFIG} ]]
      then
          echo "${NET_CONFIG} not exist"
          exit 1
      fi
  fi
}
