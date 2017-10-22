#!/bin/bash

git clone git@github.com:non117/isucon7.git
sudo cp isucon7/conf/web/etc/nginx/nginx.conf /etc/nginx/nginx.conf
sudo cp isucon7/conf/web/etc/nginx/sites-enabled/nginx.conf /etc/nginx/sites-enabled/nginx.conf
cp /home/isucon/isucon7/app/webapp/ruby/app.rb /home/isucon/isubata/webapp/ruby/app.rb
# sudo cp isucon7/conf/my.cnf /etc/my.cnf
# redis or memcached config. do something.

# restart apps
# kill -9 `cat PATH_TO_PID`
# puma -C isucon7/conf/puma.rb
# gunicorn -c isucon7/conf/gunicorn.py

sudo nginx -t && sudo systemctl restart nginx
sudo systemctl restart isubata.ruby
# sudo systemctl restart mysql
# curl localhost:3000/hoge/fuga
