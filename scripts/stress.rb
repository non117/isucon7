require 'faraday'
require 'openssl'
require 'json'
require 'parallel'

# HOW TO USE
# bundle exec ruby stress.rb

HOST = 'https://example.com'
ARGS = [
  {
    method: :get,
    path: '/',
    params: {},
  }
]

class Stress
  def initialize
    @conn = Faraday.new(HOST) do |f|
      f.adapter :httpclient
    end
  end

  def get(arg)
    start = Time.now
    res = @conn.get arg[:path], arg[:params]
    p res.body unless res.success?
    t = Time.now - start
    arg.merge({time: t, body: res.body, status: res.status})
  end

  def post(arg)
    start = Time.now
    res = @conn.post do |req|
      req.url arg[:path]
      req.body = arg[:params]
    end
    p res.body unless res.success?
    t = Time.now - start
    arg.merge({time: t, body: res.body, status: res.status})
  end

  def call(args)
    self.send(args[:method], args)
  end

  def run
    start = Time.now
    Parallel.each(1..4, in_processes: 4) do
      # NOTE: ここから各プロセス
      results = []
      Parallel.each(1..10, in_threads: 5) do
        results << call(ARGS.sample)
      end
      puts results.to_json
    end
  end
end

s = Stress.new
s.run
