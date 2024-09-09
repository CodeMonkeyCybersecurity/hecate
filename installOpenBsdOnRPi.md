Here's a condensed set of instructions for installing OpenBSD on an RPi 4B:

### Prerequisites:
1. **Supported Device:** Ensure your Raspberry Pi 4B meets the system requirements.
2. **Installation Media:** Use another machine to write the miniroot image to an SD card.
3. **Serial Console Access:** Ensure you have a serial console setup to interact with the board's firmware.

### Download OpenBSD:
1. Visit [OpenBSD FTP mirrors](https://www.openbsd.org/ftp.html) and download the appropriate files:
   - `miniroot-rpi4.img` (if available) or another miniroot suitable for your device.
   - `SHA256` and `SHA256.sig` for verification.
2. Verify the downloaded files:
   ```sh
   signify -C -p /etc/signify/openbsd-75-base.pub -x SHA256.sig <downloaded-file>
   ```

### Prepare Installation Media:
1. Write the miniroot image to an SD card:
   ```sh
   dd if=miniroot-rpi4.img of=/dev/rsdXc bs=1M
   ```
   Replace `/dev/rsdXc` with the SD card device path.

### Boot from SD Card:
1. Insert the SD card into the Raspberry Pi.
2. Connect via serial console (e.g., `cu -l cuaU0 -s 115200` on OpenBSD).
3. Boot the Raspberry Pi, ensuring it boots from the SD card.

### Start Installation:
1. At the `boot>` prompt, boot the installer:
   ```sh
   boot> set tty fb0  # If using framebuffer console
   boot> boot
   ```
2. Choose `(I)nstall` to start a fresh installation.

### Install OpenBSD:
1. **System Configuration:**
   - Set the hostname.
   - Configure network interfaces (DHCP recommended).
   - Set a root password and optionally create a user.
2. **Partition Disks:**
   - Use the suggested MBR/GPT partitioning scheme. Ensure one partition is 'OpenBSD' and another 'MSDOS' for U-Boot.
   - Accept the default filesystem layout or customize as needed.
3. **Select Installation Sets:**
   - Choose sets to install (`base75`, `comp75`, `xbase75`, etc.). Most users will select `all`.
4. **Install Bootloader:** Automatically handled by the installer.

### Final Steps:
1. Reboot the system after installation.
2. Login as `root`, create a new user account if not done during installation.
3. Review and tailor system configurations as needed by running:
   ```sh
   man afterboot
   ```

### Additional Information:
- **Packages & Ports:** Install additional software using `pkg_add` or by building from ports.
- **Documentation:** Utilize the manual pages (`man`) for system administration help.
- **Community & Support:** Refer to mailing lists and [OpenBSD documentation](https://www.openbsd.org).

This guide provides a concise overview of the OpenBSD installation process on an RPi 4B. For detailed steps and troubleshooting, refer to the official OpenBSD installation notes.
