#include "HashStorage.cuh"


HashStorage::HashStorage(const uint64_t hashLimit) : hashLimit(hashLimit) {
	hashes.reserve(hashLimit);
}

void HashStorage::addHash(uint64_t hash) {
	hashes.push_back(hash);
}

HashStorage::~HashStorage() {
}

std::vector<uint64_t> HashStorage::search(const uint64_t &hash, const uint64_t maxDistance) {
	thrust::device_vector<uint64_t> matches(hashLimit, 0);
	thrust::copy_if(
					hashes.cbegin(), hashes.cend(),
					matches.begin(),
					HammingDistanceFilter(hash, maxDistance)
			);

	//thrust::sort(matches.begin(), matches.end());

	std::vector<uint64_t> hostMatches(hashLimit);
	thrust::copy(matches.cbegin(), matches.cend(), hostMatches.begin());

	std::vector<uint64_t> results;

	results.push_back(0);
	for (int i = 0; i < hashLimit; i++) {
			uint64_t *hash = &(hostMatches[i]);
			if (*hash != 0) {
				results.push_back(*hash);
			}
		}
	return results;
}
