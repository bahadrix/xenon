/*
 ============================================================================
 Name        : CyclopsCore.cu
 Author      : Bahadir
 Version     :
 Copyright   : Allrights are alright
 Description : Compute sum of reciprocals using STL on CPU and Thrust on GPU
 ============================================================================
 */

extern "C" {
#include "XenonCore.h"

}

#include "HashStorage.cuh"
#include <vector>


static HashStorage* storage;

extern "C" {

	const uint64_t ZERO = 0;

	void initStorage(const uint64_t size) {

		storage = new HashStorage(size);

	}

	void addHash(uint64_t hash) {
		storage->addHash(hash);
	}



	Results search(const uint64_t hash, const uint64_t maxDistance) {

		std::vector<uint64_t> hashes = storage->search(hash, maxDistance);

		struct Results results;

		results.size = hashes.size();


		results.hashPtr = results.size > 0 ? &hashes[0] : &ZERO;



		return results;

	}

	uint64_t hashSize() {
		return storage->getHashes()->size();
	}



}
