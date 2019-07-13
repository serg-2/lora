# Dragino project 
In Progress

# TO enable service
sudo cp lora.service /lib/systemd/system/lora.service
sudo chmod 0644 /lib/systemd/system/lora.service
sudo chown root:root /lib/systemd/system/lora.service
sudo systemctl enable lora.service
sudo systemctl start lora.service

