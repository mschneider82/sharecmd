class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.87/sharecmd_0.0.87_Darwin_x86_64.tar.gz"
  version "0.0.87"
  sha256 "e6c82708e098b5f9f2de7d9d6414bc7a5e907e4386f93ef2400dfd26fe34aa01"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
