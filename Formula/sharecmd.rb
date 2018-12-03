class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.56/sharecmd_0.0.56_Darwin_x86_64.tar.gz"
  version "0.0.56"
  sha256 "985fea601fb670268c41df2ebb5e4375c82564a4510cd1f7caebb22db57c9a9c"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
