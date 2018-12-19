class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.65/sharecmd_0.0.65_Darwin_x86_64.tar.gz"
  version "0.0.65"
  sha256 "3aee07f8868922932d918ad72a5a1fd24fde745b96524f077082c9aa53817dce"

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
