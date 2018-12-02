class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.50/sharecmd_0.0.50_Darwin_x86_64.tar.gz"
  version "0.0.50"
  sha256 "93d2a0d22bd21585579dcf24240ef9d4271c0cbbb4eeadf5e87d3bd349d82e8d"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
